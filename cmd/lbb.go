package main

import (
	"fmt"
	"regexp"
	"strings"
)

type lbbTableEntry struct {
	pattern string
	abbrev  string
	re      *regexp.Regexp
}

var lbbCaseNamesAndInstitutionalAuthors []lbbTableEntry
var lbbStates []lbbTableEntry
var lbbCities []lbbTableEntry
var lbbTerritories []lbbTableEntry
var lbbAustralia []lbbTableEntry
var lbbCanada []lbbTableEntry
var lbbCountriesAndRegions []lbbTableEntry
var lbbPublishingTerms []lbbTableEntry
var lbbSubdivisions []lbbTableEntry

type lbbEncoder struct {
	cfg          serviceConfigFormat
	url          string
	preferCiteAs bool
	data         *genericCitation
}

func newLbbEncoder(cfg serviceConfigFormat, preferCiteAs bool) *lbbEncoder {
	e := lbbEncoder{}

	e.cfg = cfg
	e.preferCiteAs = preferCiteAs

	return &e
}

func (e *lbbEncoder) Init(url string) {
	e.url = url
}

func (e *lbbEncoder) Populate(parts citationParts) error {
	var err error

	opts := genericCitationOpts{
		verbose:        true,
		stripProtocol:  true,
		volumePrefix:   false,
		issuePrefix:    false,
		pagesPrefix:    false,
		publisherPlace: false,
	}

	if e.data, err = newGenericCitation(e.url, parts, opts); err != nil {
		return err
	}

	return nil
}

func (e *lbbEncoder) Label() string {
	return e.cfg.Label
}

func (e *lbbEncoder) ContentType() string {
	return e.cfg.ContentType
}

func (e *lbbEncoder) FileName() string {
	return ""
}

func (e *lbbEncoder) buildName(name string) string {
	return e.abbreviateNames(readingOrder(name))
}

func (e *lbbEncoder) buildNames(names []string) string {
	res := ""

	switch len(names) {
	case 0:

	case 1:
		res = e.buildName(names[0])

	case 2:
		res = fmt.Sprintf("%s & %s", e.buildName(names[0]), e.buildName(names[1]))

	default:
		res = fmt.Sprintf("%s et al.", e.buildName(names[0]))
	}

	return res
}

func (e *lbbEncoder) buildAuthors(names []string) string {
	res := e.buildNames(names)

	return res
}

func (e *lbbEncoder) buildEditors(names []string) string {
	res := e.buildNames(names)

	switch len(names) {
	case 0:

	case 1:
		res += " ed."

	default:
		res += " eds."
	}

	return res
}

func (e *lbbEncoder) buildTranslators(names []string) string {
	res := e.buildNames(names)

	switch len(names) {
	case 0:

	default:
		res += " trans."
	}

	return res
}

func (e *lbbEncoder) bookCitation() string {
	res := ""

	authors := e.buildAuthors(e.data.authors)

	if authors != "" {
		res = smallCaps(authors) + ", "
	}

	res += smallCaps(e.data.title)

	// build parenthetical piece upward

	var spaced []string

	if e.data.edition != "" {
		spaced = append(spaced, e.data.edition)
	}

	if e.data.year != 0 {
		spaced = append(spaced, fmt.Sprintf("%d", e.data.year))
	}

	var commad []string

	if s := e.buildEditors(e.data.editors); s != "" {
		commad = append(commad, s)
	}

	if s := e.buildTranslators(e.data.translators); s != "" {
		commad = append(commad, s)
	}

	if s := e.buildEditors(e.data.editors); s != "" {
		commad = append(commad, s)
	}

	if len(spaced) > 0 {
		commad = append(commad, strings.Join(spaced, " "))
	}

	if len(commad) > 0 {
		res += " ("
		res += strings.Join(commad, ", ")
		res += ")"
	}

	res += "."

	return res
}

func (e *lbbEncoder) articleCitation() string {
	res := ""

	var commad []string

	if s := e.buildAuthors(e.data.authors); s != "" {
		commad = append(commad, s)
	}

	if s := italics(mlaTitle(e.data.title)); s != "" {
		commad = append(commad, s)
	}

	if s := smallCaps(e.data.journal); s != "" {
		commad = append(commad, s)
	}

	if s := lbbDate(e.data.year, e.data.month, e.data.day); s != "" {
		commad = append(commad, s)
	}

	if e.data.pageFrom != "" {
		commad = append(commad, fmt.Sprintf("at %s", e.data.pageFrom))
	}

	res = strings.Join(commad, ", ") + "."

	return res
}

func (e *lbbEncoder) mediaCitation() string {
	res := smallCaps(e.data.title)

	// build parenthetical piece upward

	var spaced []string

	if e.data.publisher != "" {
		spaced = append(spaced, e.data.publisher)
	}

	if e.data.year != 0 {
		spaced = append(spaced, fmt.Sprintf("%d", e.data.year))
	}

	if len(spaced) > 0 {
		res += " ("
		res += strings.Join(spaced, " ")
		res += ")"
	}

	res += "."

	return res
}

func (e *lbbEncoder) Contents() (string, error) {
	if e.preferCiteAs == true && len(e.data.citeAs) > 0 {
		return strings.Join(e.data.citeAs, "\n"), nil
	}

	/*
	   "book"
	   "government_document"

	   "video"
	   "sound"

	   "journal"

	   "music"
	   "manuscript"
	   "news"
	   "thesis"
	   "map"
	   "art"
	   "generic"
	*/

	switch e.data.format {
	case "book":
		fallthrough
	case "government_document":
		return e.bookCitation(), nil

	case "sound":
		fallthrough
	case "video":
		return e.mediaCitation(), nil

	case "article":
		return e.articleCitation(), nil

	default:
		// book format is a good fallback since it uses several common generic fields.
		// this should at least generate a minimal citation, even if it's not correct.
		return e.bookCitation(), nil
	}
}

func lbbDate(y, m, d int) string {
	res := ""

	month := monthName(m)
	if len(month) > 3 {
		month = month[:3] + "."
	}

	switch {
	case y != 0 && month != "" && d != 0:
		res = fmt.Sprintf("%s %d, %d", month, d, y)

	case y != 0 && month != "":
		res = fmt.Sprintf("%s %d", month, y)

	case y != 0:
		res = fmt.Sprintf("%d", y)
	}

	return res
}

func (e *lbbEncoder) abbreviateNames(str string) string {
	res := e.abbreviateCaseNamesAndInstitutionalAuthors(str)
	res = e.abbreviateGeographicalTerms(res)
	return res
}

func (e *lbbEncoder) abbreviateCaseNamesAndInstitutionalAuthors(str string) string {
	res := str

	for i := range lbbCaseNamesAndInstitutionalAuthors {
		entry := &lbbCaseNamesAndInstitutionalAuthors[i]

		res = entry.re.ReplaceAllString(res, entry.abbrev)
	}

	return res
}

func (e *lbbEncoder) abbreviateGeographicalTerms(str string) string {
	res := str

	for i := range lbbStates {
		entry := &lbbStates[i]

		res = entry.re.ReplaceAllString(res, entry.abbrev)
	}

	for i := range lbbCities {
		entry := &lbbCities[i]

		res = entry.re.ReplaceAllString(res, entry.abbrev)
	}

	for i := range lbbTerritories {
		entry := &lbbTerritories[i]

		res = entry.re.ReplaceAllString(res, entry.abbrev)
	}

	for i := range lbbAustralia {
		entry := &lbbAustralia[i]

		res = entry.re.ReplaceAllString(res, entry.abbrev)
	}

	for i := range lbbCanada {
		entry := &lbbCanada[i]

		res = entry.re.ReplaceAllString(res, entry.abbrev)
	}

	for i := range lbbCountriesAndRegions {
		entry := &lbbCountriesAndRegions[i]

		res = entry.re.ReplaceAllString(res, entry.abbrev)
	}

	return res
}

func init() {
	lbbCaseNamesAndInstitutionalAuthors = []lbbTableEntry{
		{pattern: "Academ(ic|y)", abbrev: "Acad."},
		{pattern: "Account(ant|ing|ancy)", abbrev: "Acct."},
		{pattern: "Administrat(ive|ion)", abbrev: "Admin."},
		{pattern: "Administrat(or|rix)", abbrev: "Adm’(r|x)"},
		{pattern: "Advertising", abbrev: "Advert."},
		{pattern: "Advoca(te|cy)", abbrev: "Advoc."},
		{pattern: "Affair", abbrev: "Aff."},
		{pattern: "Africa(|n)", abbrev: "Afr."},
		{pattern: "Agricultur(e|al)", abbrev: "Agric."},
		{pattern: "Alliance", abbrev: "All."},
		{pattern: "Alternative", abbrev: "Alt."},
		{pattern: "America(|n)", abbrev: "Am."},
		{pattern: "Ancestry", abbrev: "Anc."},
		{pattern: "and", abbrev: "&"},
		{pattern: "Annual", abbrev: "Ann."},
		{pattern: "Appellate", abbrev: "App."},
		{pattern: "Arbitrat(ion|or)", abbrev: "Arb."},
		{pattern: "Artificial Intelligence", abbrev: "A.I."},
		{pattern: "Associate", abbrev: "Assoc."},
		{pattern: "Association", abbrev: "Ass'n"},
		{pattern: "Atlantic", abbrev: "Atl."},
		{pattern: "Attorney", abbrev: "Att'y"},
		{pattern: "Authority", abbrev: "Auth."},
		{pattern: "Automo(bile|tive)", abbrev: "Auto."},
		{pattern: "Avenue", abbrev: "Ave."},
		{pattern: "Bankruptcy", abbrev: "Bankr."},
		{pattern: "Behavior(|al)", abbrev: "Behav."},
		{pattern: "Board", abbrev: "Bd."},
		{pattern: "British", abbrev: "Brit."},
		{pattern: "Broadcast(er|ing)", abbrev: "Broad."},
		{pattern: "Building", abbrev: "Bldg."},
		{pattern: "Bulletin", abbrev: "Bull."},
		{pattern: "Business(|es)", abbrev: "Bus."},
		{pattern: "Capital", abbrev: "Cap."},
		{pattern: "Casualt(y|ies)", abbrev: "Cas."},
		{pattern: "Catholic", abbrev: "Cath."},
		{pattern: "Cent(er|re)", abbrev: "Ctr."},
		{pattern: "Central", abbrev: "Cent."},
		{pattern: "Chemical", abbrev: "Chem."},
		{pattern: "Children", abbrev: "Child."},
		{pattern: "Chronicle", abbrev: "Chron."},
		{pattern: "Circuit", abbrev: "Cir."},
		{pattern: "Civil", abbrev: "Civ."},
		{pattern: "Civil Libert(y|ies)", abbrev: "C.L."},
		{pattern: "Civil Rights", abbrev: "C.R."},
		{pattern: "Coalition", abbrev: "Coal."},
		{pattern: "College", abbrev: "Coll."},
		{pattern: "Commentary", abbrev: "Comment."},
		{pattern: "Commerc(e|ial)", abbrev: "Com."},
		{pattern: "Commission", abbrev: "Comm'n"},
		{pattern: "Commissioner", abbrev: "Comm'r"},
		{pattern: "Committee", abbrev: "Comm."},
		{pattern: "Communication", abbrev: "Commc'n"},
		{pattern: "Community", abbrev: "Cmty."},
		{pattern: "Company", abbrev: "Co."},
		{pattern: "Comparative", abbrev: "Compar."},
		{pattern: "Compensation", abbrev: "Comp."},
		{pattern: "Computer", abbrev: "Comput."},
		{pattern: "Condominium", abbrev: "Condo."},
		{pattern: "Conference", abbrev: "Conf."},
		{pattern: "Congress(|ional)", abbrev: "Cong."},
		{pattern: "Consolidated", abbrev: "Consol."},
		{pattern: "Constitution(|al)", abbrev: "Const."},
		{pattern: "Construction", abbrev: "Constr."},
		{pattern: "Contemporary", abbrev: "Contemp."},
		{pattern: "Continental", abbrev: "Cont'l"},
		{pattern: "Contract", abbrev: "Cont."},
		{pattern: "Conveyance(|r)", abbrev: "Conv."},
		{pattern: "Cooperat(ion|ive)", abbrev: "Coop."},
		{pattern: "Corporat(e|ion)", abbrev: "Corp."},
		{pattern: "Correction(s|al)", abbrev: "Corr."},
		{pattern: "Cosmetic", abbrev: "Cosm."},
		{pattern: "Counsel(or|ors|or's)", abbrev: "Couns."},
		{pattern: "County", abbrev: "Cnty."},
		{pattern: "Court", abbrev: "Ct."},
		{pattern: "Criminal", abbrev: "Crim."},
		{pattern: "Defen(d|der|se)", abbrev: "Def."},
		{pattern: "Delinquen(t|cy)", abbrev: "Delinq."},
		{pattern: "Department", abbrev: "Dep't"},
		{pattern: "Detention", abbrev: "Det."},
		{pattern: "Develop(er|ment)", abbrev: "Dev."},
		{pattern: "Digest", abbrev: "Dig."},
		{pattern: "Digital", abbrev: "Digit."},
		{pattern: "Diplomacy", abbrev: "Dipl."},
		{pattern: "Director", abbrev: "Dir."},
		{pattern: "Discount", abbrev: "Disc."},
		{pattern: "Dispute", abbrev: "Disp."},
		{pattern: "Distribut(or|ing|ion)", abbrev: "Distrib."},
		{pattern: "District", abbrev: "Dist."},
		{pattern: "Division", abbrev: "Div."},
		{pattern: "Doctor", abbrev: "Dr."},
		{pattern: "East(|ern)", abbrev: "E."},
		{pattern: "Econom(ic|ical|ics|y)", abbrev: "Econ."},
		{pattern: "Editor(|ial)", abbrev: "Ed."},
		{pattern: "Education(|al)", abbrev: "Educ."},
		{pattern: "Electr(ic|ical|icity|onic)", abbrev: "Elec."},
		{pattern: "Employ(ee|er|ment)", abbrev: "Emp."},
		{pattern: "Enforcement", abbrev: "Enf't"},
		{pattern: "Engineer", abbrev: "Eng'r"},
		{pattern: "Engineering", abbrev: "Eng'g"},
		{pattern: "English", abbrev: "Eng."},
		{pattern: "Enterprise", abbrev: "Enter."},
		{pattern: "Entertainment", abbrev: "Ent."},
		{pattern: "Environment(|al)", abbrev: "Env't"},
		{pattern: "Equality", abbrev: "Equal."},
		{pattern: "Equipment", abbrev: "Equip."},
		{pattern: "Estate", abbrev: "Est."},
		{pattern: "Europe(|an)", abbrev: "Eur."},
		{pattern: "Examiner", abbrev: "Exam'r"},
		{pattern: "Exchange", abbrev: "Exch."},
		{pattern: "Executive", abbrev: "Exec."},
		{pattern: "Execut(or|rix)", abbrev: "Ex'(r|x)"},
		{pattern: "Explorat(ion|ory)", abbrev: "Expl."},
		{pattern: "Export(er|ation)", abbrev: "Exp."},
		{pattern: "Faculty", abbrev: "Fac."},
		{pattern: "Family", abbrev: "Fam."},
		{pattern: "Federal", abbrev: "Fed."},
		{pattern: "Federation", abbrev: "Fed'n"},
		{pattern: "Fidelity", abbrev: "Fid."},
		{pattern: "Financ(e|ial|ing)", abbrev: "Fin."},
		{pattern: "Fortnightly", abbrev: "Fort."},
		{pattern: "Forum", abbrev: "F."},
		{pattern: "Foundation", abbrev: "Found."},
		{pattern: "General", abbrev: "Gen."},
		{pattern: "Global", abbrev: "Glob."},
		{pattern: "Government", abbrev: "Gov't"},
		{pattern: "Group", abbrev: "Grp."},
		{pattern: "Guarant(y|or)", abbrev: "Guar."},
		{pattern: "Hispanic", abbrev: "Hisp."},
		{pattern: "Histor(ical|y)", abbrev: "Hist."},
		{pattern: "Hospital(|ity)", abbrev: "Hosp."},
		{pattern: "Housing", abbrev: "Hous."},
		{pattern: "Human", abbrev: "Hum."},
		{pattern: "Humanity", abbrev: "Human."},
		{pattern: "Immigration", abbrev: "Immigr."},
		{pattern: "Import(er|ation)", abbrev: "Imp."},
		{pattern: "Incorporated", abbrev: "Inc."},
		{pattern: "Indemnity", abbrev: "Indem."},
		{pattern: "Independen(ce|t)", abbrev: "Indep."},
		{pattern: "Industr(y|ial|ies)", abbrev: "Indus."},
		{pattern: "Inequality", abbrev: "Ineq."},
		{pattern: "Information", abbrev: "Info."},
		{pattern: "Injury", abbrev: "Inj."},
		{pattern: "Institut(e|ion)", abbrev: "Inst."},
		{pattern: "Insurance", abbrev: "Ins."},
		{pattern: "Intellectual", abbrev: "Intell."},
		{pattern: "Intelligence", abbrev: "Intel."},
		{pattern: "Interdisciplinary", abbrev: "Interdisc."},
		{pattern: "Interest", abbrev: "Int."},
		{pattern: "International", abbrev: "Int'l"},
		{pattern: "Invest(ment|or)", abbrev: "Inv."},
		{pattern: "Journal(|s)", abbrev: "J."},
		{pattern: "Judicial", abbrev: "Jud."},
		{pattern: "Juridical", abbrev: "Jurid."},
		{pattern: "Jurisprudence", abbrev: "Juris."},
		{pattern: "Justice", abbrev: "Just."},
		{pattern: "Juvenile", abbrev: "Juv."},
		{pattern: "Labor", abbrev: "Lab."},
		{pattern: "Laboratory", abbrev: "Lab'y"},
		{pattern: "Law(|s)", abbrev: "L."},
		{pattern: "Lawyer", abbrev: "Law."},
		{pattern: "Legislat(ion|ive)", abbrev: "Legis."},
		{pattern: "Liability", abbrev: "Liab."},
		{pattern: "Librar(y|ian)", abbrev: "Libr."},
		{pattern: "Limited", abbrev: "Ltd."},
		{pattern: "Litigation", abbrev: "Litig."},
		{pattern: "Local", abbrev: "Loc."},
		{pattern: "Machine(|ry)", abbrev: "Mach."},
		{pattern: "Magazine", abbrev: "Mag."},
		{pattern: "Maintenance", abbrev: "Maint."},
		{pattern: "Management", abbrev: "Mgmt."},
		{pattern: "Manufacturer", abbrev: "Mfr."},
		{pattern: "Manufacturing", abbrev: "Mfg."},
		{pattern: "Maritime", abbrev: "Mar."},
		{pattern: "Market", abbrev: "Mkt."},
		{pattern: "Marketing", abbrev: "Mktg."},
		{pattern: "Matrimonial", abbrev: "Matrim."},
		{pattern: "Mechanic(|al)", abbrev: "Mech."},
		{pattern: "Medic(al|inal|ine)", abbrev: "Med."},
		{pattern: "Memorial", abbrev: "Mem'l"},
		{pattern: "Merchan(t|dise|dising)", abbrev: "Merch."},
		{pattern: "Metropolitan", abbrev: "Metro."},
		{pattern: "Military", abbrev: "Mil."},
		{pattern: "Mineral", abbrev: "Min."},
		{pattern: "Modern", abbrev: "Mod."},
		{pattern: "Mortgage", abbrev: "Mortg."},
		{pattern: "Municipal(|ity)", abbrev: "Mun."},
		{pattern: "Mutual", abbrev: "Mut."},
		{pattern: "National", abbrev: "Nat'l"},
		{pattern: "Nationality", abbrev: "Nat'y"},
		{pattern: "Natural", abbrev: "Nat."},
		{pattern: "Negligence", abbrev: "Negl."},
		{pattern: "Negotiat(ion|or)", abbrev: "Negot."},
		{pattern: "Newsletter", abbrev: "Newsl."},
		{pattern: "North(|ern)", abbrev: "N."},
		{pattern: "Northeast(|ern)", abbrev: "Ne."},
		{pattern: "Northwest(|ern)", abbrev: "Nw."},
		{pattern: "Number", abbrev: "No."},
		{pattern: "Offic(e|ial)", abbrev: "Off."},
		{pattern: "Opinion", abbrev: "Op."},
		{pattern: "Order", abbrev: "Ord."},
		{pattern: "Organiz(ation|ing)", abbrev: "Org."},
		{pattern: "Pacific", abbrev: "Pac."},
		{pattern: "Parish", abbrev: "Par."},
		{pattern: "Partnership", abbrev: "P'ship"},
		{pattern: "Patent", abbrev: "Pat."},
		{pattern: "Person(al|nel)", abbrev: "Pers."},
		{pattern: "Perspective", abbrev: "Persp."},
		{pattern: "Pharmaceutic(|al)", abbrev: "Pharm."},
		{pattern: "Philosoph(ical|y)", abbrev: "Phil."},
		{pattern: "Planning", abbrev: "Plan."},
		{pattern: "Policy", abbrev: "Pol'y"},
		{pattern: "Politic(al|s)", abbrev: "Pol."},
		{pattern: "Practi(cal|ce|titioner)", abbrev: "Prac."},
		{pattern: "Preserv(e|ation)", abbrev: "Pres."},
		{pattern: "Priva(cy|te)", abbrev: "Priv."},
		{pattern: "Probat(e|ion)", abbrev: "Prob."},
		{pattern: "Problems", abbrev: "Probs."},
		{pattern: "Proce(edings|dure)", abbrev: "Proc."},
		{pattern: "Product(|ion)", abbrev: "Prod."},
		{pattern: "Profession(|al)", abbrev: "Pro."},
		{pattern: "Property", abbrev: "Prop."},
		{pattern: "Protection", abbrev: "Prot."},
		{pattern: "Psycholog(ical|ist|y)", abbrev: "Psych."},
		{pattern: "Public", abbrev: "Pub."},
		{pattern: "Publication", abbrev: "Publ'n"},
		{pattern: "Publishing", abbrev: "Publ'g"},
		{pattern: "Quarterly", abbrev: "Q."},
		{pattern: "Railroad", abbrev: "R.R."},
		{pattern: "Railway", abbrev: "Ry."},
		{pattern: "Record", abbrev: "Rec."},
		{pattern: "Referee", abbrev: "Ref."},
		{pattern: "Refin(ing|ement)", abbrev: "Refin."},
		{pattern: "Regional", abbrev: "Reg'l"},
		{pattern: "Register", abbrev: "Reg."},
		{pattern: "Regulat(ion|or|ory)", abbrev: "Regul."},
		{pattern: "Rehabilitat(ion|ive)", abbrev: "Rehab."},
		{pattern: "Relation", abbrev: "Rel."},
		{pattern: "Report(|er)", abbrev: "Rep."},
		{pattern: "Reproduct(ion|ive)", abbrev: "Reprod."},
		{pattern: "Research", abbrev: "Rsch."},
		{pattern: "Reserv(ation|e)", abbrev: "Rsrv."},
		{pattern: "Resolution", abbrev: "Resol."},
		{pattern: "Resource(|s)", abbrev: "Res."},
		{pattern: "Responsibility", abbrev: "Resp."},
		{pattern: "Restaurant", abbrev: "Rest."},
		{pattern: "Retirement", abbrev: "Ret."},
		{pattern: "Review|Revista", abbrev: "Rev."},
		{pattern: "Rights", abbrev: "Rts."},
		{pattern: "Road", abbrev: "Rd."},
		{pattern: "Savings", abbrev: "Sav."},
		{pattern: "School", abbrev: "Sch."},
		{pattern: "Scien(ce|tific)", abbrev: "Sci."},
		{pattern: "Scottish", abbrev: "Scot."},
		{pattern: "Secretary", abbrev: "Sec'y"},
		{pattern: "Securit(y|ies)", abbrev: "Sec."},
		{pattern: "Sentencing", abbrev: "Sent'g"},
		{pattern: "Service", abbrev: "Serv."},
		{pattern: "Shareholder|Stockholder", abbrev: "S'holder"},
		{pattern: "Social", abbrev: "Soc."},
		{pattern: "Society", abbrev: "Soc'y"},
		{pattern: "Sociolog(ical|y)", abbrev: "Socio."},
		{pattern: "Solicitor", abbrev: "Solic."},
		{pattern: "Solution", abbrev: "Sol."},
		{pattern: "South(|ern)", abbrev: "S."},
		{pattern: "Southeast(|ern)", abbrev: "Se."},
		{pattern: "Southwest(|ern)", abbrev: "Sw."},
		{pattern: "Statistic(s|al)", abbrev: "Stat."},
		{pattern: "Steamship(|s)", abbrev: "S.S."},
		{pattern: "Street", abbrev: "St."},
		{pattern: "Studies", abbrev: "Stud."},
		{pattern: "Subcommittee", abbrev: "Subcomm."},
		{pattern: "Supreme Court", abbrev: "Sup. Ct."},
		{pattern: "Surety", abbrev: "Sur."},
		{pattern: "Survey", abbrev: "Surv."},
		{pattern: "Symposium", abbrev: "Symp."},
		{pattern: "System(|s)", abbrev: "Sys."},
		{pattern: "Taxation", abbrev: "Tax'n"},
		{pattern: "Teacher", abbrev: "Tchr."},
		{pattern: "Techn(ical|ique|ology|ological)", abbrev: "Tech."},
		{pattern: "Telecommunication", abbrev: "Telecomm."},
		{pattern: "Tele(phone|graph)", abbrev: "Tel."},
		{pattern: "Temporary", abbrev: "Temp."},
		{pattern: "Township", abbrev: "Twp."},
		{pattern: "Transcontinental", abbrev: "Transcon."},
		{pattern: "Transnational", abbrev: "Transnat'l"},
		{pattern: "Transport(|ation)", abbrev: "Transp."},
		{pattern: "Tribune", abbrev: "Trib."},
		{pattern: "Trust(|ee)", abbrev: "Tr."},
		{pattern: "Turnpike", abbrev: "Tpk."},
		{pattern: "Uniform", abbrev: "Unif."},
		{pattern: "United States", abbrev: "U.S."},
		{pattern: "University", abbrev: "Univ."},
		{pattern: "Urban", abbrev: "Urb."},
		{pattern: "Utility", abbrev: "Util."},
		{pattern: "Village", abbrev: "Vill."},
		{pattern: "Week", abbrev: "Wk."},
		{pattern: "Weekly", abbrev: "Wkly."},
		{pattern: "West(|ern)", abbrev: "W."},
		{pattern: "Year(| )book", abbrev: "Y.B."},
	}

	for i := range lbbCaseNamesAndInstitutionalAuthors {
		entry := &lbbCaseNamesAndInstitutionalAuthors[i]

		pattern := fmt.Sprintf(`(?i)\b%s\b`, entry.pattern)
		entry.re = regexp.MustCompile(pattern)
	}

	lbbStates = []lbbTableEntry{
		{pattern: "Alabama", abbrev: "Ala."},
		{pattern: "Alaska", abbrev: "Alaska"},
		{pattern: "Arizona", abbrev: "Ariz."},
		{pattern: "Arkansas", abbrev: "Ark."},
		{pattern: "California", abbrev: "Cal."},
		{pattern: "Colorado", abbrev: "Colo."},
		{pattern: "Connecticut", abbrev: "Conn."},
		{pattern: "Delaware", abbrev: "Del."},
		{pattern: "Florida", abbrev: "Fla."},
		{pattern: "Georgia", abbrev: "Ga."},
		{pattern: "Hawaii", abbrev: "Haw."},
		{pattern: "Idaho", abbrev: "Idaho"},
		{pattern: "Illinois", abbrev: "Ill."},
		{pattern: "Indiana", abbrev: "Ind."},
		{pattern: "Iowa", abbrev: "Iowa"},
		{pattern: "Kansas", abbrev: "Kan."},
		{pattern: "Kentucky", abbrev: "Ky."},
		{pattern: "Louisiana", abbrev: "La."},
		{pattern: "Maine", abbrev: "Me."},
		{pattern: "Maryland", abbrev: "Md."},
		{pattern: "Massachusetts", abbrev: "Mass."},
		{pattern: "Michigan", abbrev: "Mich."},
		{pattern: "Minnesota", abbrev: "Minn."},
		{pattern: "Mississippi", abbrev: "Miss."},
		{pattern: "Missouri", abbrev: "Mo."},
		{pattern: "Montana", abbrev: "Mont."},
		{pattern: "Nebraska", abbrev: "Neb."},
		{pattern: "Nevada", abbrev: "Nev."},
		{pattern: "New Hampshire", abbrev: "N.H."},
		{pattern: "New Jersey", abbrev: "N.J."},
		{pattern: "New Mexico", abbrev: "N.M."},
		{pattern: "New York", abbrev: "N.Y."},
		{pattern: "North Carolina", abbrev: "N.C."},
		{pattern: "North Dakota", abbrev: "N.D."},
		{pattern: "Ohio", abbrev: "Ohio"},
		{pattern: "Oklahoma", abbrev: "Okla."},
		{pattern: "Oregon", abbrev: "Or."},
		{pattern: "Pennsylvania", abbrev: "Pa."},
		{pattern: "Rhode Island", abbrev: "R.I."},
		{pattern: "South Carolina", abbrev: "S.C."},
		{pattern: "South Dakota", abbrev: "S.D."},
		{pattern: "Tennessee", abbrev: "Tenn."},
		{pattern: "Texas", abbrev: "Tex."},
		{pattern: "Utah", abbrev: "Utah"},
		{pattern: "Vermont", abbrev: "Vt."},
		{pattern: "Virginia", abbrev: "Va."},
		{pattern: "Washington", abbrev: "Wash."},
		{pattern: "West Virginia", abbrev: "W. Va."},
		{pattern: "Wisconsin", abbrev: "Wis."},
		{pattern: "Wyoming", abbrev: "Wyo."},
	}

	for i := range lbbStates {
		entry := &lbbStates[i]

		pattern := fmt.Sprintf(`(?i)\b%s\b`, entry.pattern)
		entry.re = regexp.MustCompile(pattern)
	}

	lbbCities = []lbbTableEntry{
		{pattern: "Baltimore", abbrev: "Balt."},
		{pattern: "Boston", abbrev: "Bos."},
		{pattern: "Chicago", abbrev: "Chi."},
		{pattern: "Dallas", abbrev: "Dall."},
		{pattern: "District of Columbia", abbrev: "D.C."},
		{pattern: "Houston", abbrev: "Hous."},
		{pattern: "Los Angeles", abbrev: "L.A."},
		{pattern: "Miami", abbrev: "Mia."},
		{pattern: "New York", abbrev: "N.Y.C."},
		{pattern: "Philadelphia", abbrev: "Phila."},
		{pattern: "Phoenix", abbrev: "Phx."},
		{pattern: "San Francisco", abbrev: "S.F."},
	}

	for i := range lbbCities {
		entry := &lbbCities[i]

		pattern := fmt.Sprintf(`(?i)\b%s\b`, entry.pattern)
		entry.re = regexp.MustCompile(pattern)
	}

	lbbTerritories = []lbbTableEntry{
		{pattern: "American Samoa", abbrev: "Am. Sam."},
		{pattern: "Guam", abbrev: "Guam"},
		{pattern: "Northern Mariana Islands", abbrev: "N. Mar. I."},
		{pattern: "Puerto Rico", abbrev: "P.R."},
		{pattern: "Virgin Islands", abbrev: "V.I."},
	}

	for i := range lbbTerritories {
		entry := &lbbTerritories[i]

		pattern := fmt.Sprintf(`(?i)\b%s\b`, entry.pattern)
		entry.re = regexp.MustCompile(pattern)
	}

	lbbAustralia = []lbbTableEntry{
		{pattern: "Australian Capital Territory", abbrev: "Austl. Cap. Terr."},
		{pattern: "New South Wales", abbrev: "N.S.W."},
		{pattern: "Northern Territory", abbrev: "N. Terr."},
		{pattern: "Queensland", abbrev: "Queensl."},
		{pattern: "South Australia", abbrev: "S. Austl."},
		{pattern: "Tasmania", abbrev: "Tas."},
		{pattern: "Victoria", abbrev: "Vict."},
		{pattern: "Western Australia", abbrev: "W. Austl."},
	}

	for i := range lbbAustralia {
		entry := &lbbAustralia[i]

		pattern := fmt.Sprintf(`(?i)\b%s\b`, entry.pattern)
		entry.re = regexp.MustCompile(pattern)
	}

	lbbCanada = []lbbTableEntry{
		{pattern: "Alberta", abbrev: "Alta."},
		{pattern: "British Columbia", abbrev: "B.C."},
		{pattern: "Manitoba", abbrev: "Man."},
		{pattern: "New Brunswick", abbrev: "N.B."},
		{pattern: "Newfoundland & Labrador", abbrev: "Nfld."},
		{pattern: "Northwest Territories", abbrev: "N.W.T."},
		{pattern: "Nova Scotia", abbrev: "N.S."},
		{pattern: "Nunavut", abbrev: "Nun."},
		{pattern: "Ontario", abbrev: "Ont."},
		{pattern: "Prince Edward Island", abbrev: "P.E.I."},
		{pattern: "Québec", abbrev: "Que."},
		{pattern: "Saskatchewan", abbrev: "Sask."},
		{pattern: "Yukon", abbrev: "Yukon"},
	}

	for i := range lbbCanada {
		entry := &lbbCanada[i]

		pattern := fmt.Sprintf(`(?i)\b%s\b`, entry.pattern)
		entry.re = regexp.MustCompile(pattern)
	}

	lbbCountriesAndRegions = []lbbTableEntry{
		{pattern: "Afghanistan", abbrev: "Afg."},
		{pattern: "Africa", abbrev: "Afr."},
		{pattern: "Albania", abbrev: "Alb."},
		{pattern: "Algeria", abbrev: "Alg."},
		{pattern: "Andorra", abbrev: "Andorra"},
		{pattern: "Angola", abbrev: "Angl."},
		{pattern: "Anguilla", abbrev: "Anguilla"},
		{pattern: "Antarctica", abbrev: "Antarctica"},
		{pattern: "Antigua & Barbuda", abbrev: "Ant. & Barb."},
		{pattern: "Argentina", abbrev: "Arg."},
		{pattern: "Armenia", abbrev: "Arm."},
		{pattern: "Asia", abbrev: "Asia"},
		{pattern: "Australia", abbrev: "Austl."},
		{pattern: "Austria", abbrev: "Austria"},
		{pattern: "Azerbaijan", abbrev: "Azer."},
		{pattern: "Bahamas", abbrev: "Bah."},
		{pattern: "Bahrain", abbrev: "Bahr."},
		{pattern: "Bangladesh", abbrev: "Bangl."},
		{pattern: "Barbados", abbrev: "Barb."},
		{pattern: "Belarus", abbrev: "Belr."},
		{pattern: "Belgium", abbrev: "Belg."},
		{pattern: "Belize", abbrev: "Belize"},
		{pattern: "Benin", abbrev: "Benin"},
		{pattern: "Bermuda", abbrev: "Berm."},
		{pattern: "Bhutan", abbrev: "Bhutan"},
		{pattern: "Bolivia", abbrev: "Bol."},
		{pattern: "Bosnia & Herzegovina", abbrev: "Bosn. & Herz."},
		{pattern: "Botswana", abbrev: "Bots."},
		{pattern: "Brazil", abbrev: "Braz."},
		{pattern: "Brunei", abbrev: "Brunei"},
		{pattern: "Bulgaria", abbrev: "Bulg."},
		{pattern: "Burkina Faso", abbrev: "Burk. Faso"},
		{pattern: "Burundi", abbrev: "Burundi"},
		{pattern: "Cambodia", abbrev: "Cambodia"},
		{pattern: "Cameroon", abbrev: "Cameroon"},
		{pattern: "Canada", abbrev: "Can."},
		{pattern: "Cape Verde", abbrev: "Cape Verde"},
		{pattern: "Cayman Islands", abbrev: "Cayman Is."},
		{pattern: "Central African Republic", abbrev: "Cent. Afr. Rep."},
		{pattern: "Chad", abbrev: "Chad"},
		{pattern: "Chile", abbrev: "Chile"},
		{pattern: "China, People’s Republic of", abbrev: "China"},
		{pattern: "Colombia", abbrev: "Colom."},
		{pattern: "Comoros", abbrev: "Comoros"},
		{pattern: "Congo, Democratic Republic of the", abbrev: "Dem. Rep. Congo"},
		{pattern: "Congo, Republic of the", abbrev: "Congo"},
		{pattern: "Costa Rica", abbrev: "Costa Rica"},
		{pattern: "Côte d’Ivoire", abbrev: "Côte d’Ivoire"},
		{pattern: "Croatia", abbrev: "Croat."},
		{pattern: "Cuba", abbrev: "Cuba"},
		{pattern: "Cyprus", abbrev: "Cyprus"},
		{pattern: "Czech Republic", abbrev: "Czech"},
		{pattern: "Denmark", abbrev: "Den."},
		{pattern: "Djibouti", abbrev: "Djib."},
		{pattern: "Dominica", abbrev: "Dominica"},
		{pattern: "Dominican Republic", abbrev: "Dom. Rep."},
		{pattern: "Ecuador", abbrev: "Ecuador"},
		{pattern: "Egypt", abbrev: "Egypt"},
		{pattern: "El Salvador", abbrev: "El Sal."},
		{pattern: "England", abbrev: "Eng."},
		{pattern: "Equatorial Guinea", abbrev: "Eq. Guinea"},
		{pattern: "Eritrea", abbrev: "Eri."},
		{pattern: "Estonia", abbrev: "Est."},
		{pattern: "Ethiopia", abbrev: "Eth."},
		{pattern: "Europe", abbrev: "Eur."},
		{pattern: "Falkland Islands", abbrev: "Falkland Is."},
		{pattern: "Fiji", abbrev: "Fiji"},
		{pattern: "Finland", abbrev: "Fin."},
		{pattern: "France", abbrev: "Fr."},
		{pattern: "Gabon", abbrev: "Gabon"},
		{pattern: "Gambia", abbrev: "Gam."},
		{pattern: "Georgia", abbrev: "Geor."},
		{pattern: "Germany", abbrev: "Ger."},
		{pattern: "Ghana", abbrev: "Ghana"},
		{pattern: "Gibraltar", abbrev: "Gib."},
		{pattern: "Great Britain", abbrev: "Gr. Brit."},
		{pattern: "Greece", abbrev: "Greece"},
		{pattern: "Greenland", abbrev: "Green."},
		{pattern: "Grenada", abbrev: "Gren."},
		{pattern: "Guadeloupe", abbrev: "Guad."},
		{pattern: "Guatemala", abbrev: "Guat."},
		{pattern: "Guinea", abbrev: "Guinea"},
		{pattern: "Guinea-Bissau", abbrev: "Guinea-Bissau"},
		{pattern: "Guyana", abbrev: "Guy."},
		{pattern: "Haiti", abbrev: "Haiti"},
		{pattern: "Honduras", abbrev: "Hond."},
		{pattern: "Hong Kong", abbrev: "H.K."},
		{pattern: "Hungary", abbrev: "Hung."},
		{pattern: "Iceland", abbrev: "Ice."},
		{pattern: "India", abbrev: "India"},
		{pattern: "Indonesia", abbrev: "Indon."},
		{pattern: "Iran", abbrev: "Iran"},
		{pattern: "Iraq", abbrev: "Iraq"},
		{pattern: "Ireland", abbrev: "Ir."},
		{pattern: "Israel", abbrev: "Isr."},
		{pattern: "Italy", abbrev: "It."},
		{pattern: "Jamaica", abbrev: "Jam."},
		{pattern: "Japan", abbrev: "Japan"},
		{pattern: "Jordan", abbrev: "Jordan"},
		{pattern: "Kazakhstan", abbrev: "Kaz."},
		{pattern: "Kenya", abbrev: "Kenya"},
		{pattern: "Kiribati", abbrev: "Kiribati"},
		{pattern: "Korea, North", abbrev: "N. Kor."},
		{pattern: "Korea, South", abbrev: "S. Kor."},
		{pattern: "Kosovo", abbrev: "Kos."},
		{pattern: "Kuwait", abbrev: "Kuwait"},
		{pattern: "Kyrgyzstan", abbrev: "Kyrg."},
		{pattern: "Laos", abbrev: "Laos"},
		{pattern: "Latvia", abbrev: "Lat."},
		{pattern: "Lebanon", abbrev: "Leb."},
		{pattern: "Lesotho", abbrev: "Lesotho"},
		{pattern: "Liberia", abbrev: "Liber."},
		{pattern: "Libya", abbrev: "Libya"},
		{pattern: "Liechtenstein", abbrev: "Liech."},
		{pattern: "Lithuania", abbrev: "Lith."},
		{pattern: "Luxembourg", abbrev: "Lux."},
		{pattern: "Macau", abbrev: "Mac."},
		{pattern: "Macedonia", abbrev: "Maced."},
		{pattern: "Madagascar", abbrev: "Madag."},
		{pattern: "Malawi", abbrev: "Malawi"},
		{pattern: "Malaysia", abbrev: "Malay."},
		{pattern: "Maldives", abbrev: "Maldives"},
		{pattern: "Mali", abbrev: "Mali"},
		{pattern: "Malta", abbrev: "Malta"},
		{pattern: "Marshall Islands", abbrev: "Marsh. Is."},
		{pattern: "Martinique", abbrev: "Mart."},
		{pattern: "Mauritania", abbrev: "Mauritania"},
		{pattern: "Mauritius", abbrev: "Mauritius"},
		{pattern: "Mexico", abbrev: "Mex."},
		{pattern: "Micronesia", abbrev: "Micr."},
		{pattern: "Moldova", abbrev: "Mold."},
		{pattern: "Monaco", abbrev: "Monaco"},
		{pattern: "Mongolia", abbrev: "Mong."},
		{pattern: "Montenegro", abbrev: "Montenegro"},
		{pattern: "Montserrat", abbrev: "Montserrat"},
		{pattern: "Morocco", abbrev: "Morocco"},
		{pattern: "Mozambique", abbrev: "Mozam."},
		{pattern: "Myanmar", abbrev: "Myan."},
		{pattern: "Namibia", abbrev: "Namib."},
		{pattern: "Nauru", abbrev: "Nauru"},
		{pattern: "Nepal", abbrev: "Nepal"},
		{pattern: "Netherlands", abbrev: "Neth."},
		{pattern: "New Zealand", abbrev: "N.Z."},
		{pattern: "Nicaragua", abbrev: "Nicar."},
		{pattern: "Niger", abbrev: "Niger"},
		{pattern: "Nigeria", abbrev: "Nigeria"},
		{pattern: "North America", abbrev: "N. Am."},
		{pattern: "Northern Ireland", abbrev: "N. Ir."},
		{pattern: "Norway", abbrev: "Nor."},
		{pattern: "Oman", abbrev: "Oman"},
		{pattern: "Pakistan", abbrev: "Pak."},
		{pattern: "Palau", abbrev: "Palau"},
		{pattern: "Panama", abbrev: "Pan."},
		{pattern: "Papua New Guinea", abbrev: "Papua N.G."},
		{pattern: "Paraguay", abbrev: "Para."},
		{pattern: "Peru", abbrev: "Peru"},
		{pattern: "Philippines", abbrev: "Phil."},
		{pattern: "Pitcairn Island", abbrev: "Pitcairn Is."},
		{pattern: "Poland", abbrev: "Pol."},
		{pattern: "Portugal", abbrev: "Port."},
		{pattern: "Qatar", abbrev: "Qatar"},
		{pattern: "Réunion", abbrev: "Réunion"},
		{pattern: "Romania", abbrev: "Rom."},
		{pattern: "Russia", abbrev: "Russ."},
		{pattern: "Rwanda", abbrev: "Rwanda"},
		{pattern: "Saint Helena", abbrev: "St. Helena"},
		{pattern: "Saint Kitts & Nevis", abbrev: "St. Kitts & Nevis"},
		{pattern: "Saint Lucia", abbrev: "St. Lucia"},
		{pattern: "Saint Vincent & the Grenadines", abbrev: "St. Vincent"},
		{pattern: "Samoa", abbrev: "Samoa"},
		{pattern: "San Marino", abbrev: "San Marino"},
		{pattern: "São Tomé and Príncipe", abbrev: "São Tomé & Príncipe"},
		{pattern: "Saudi Arabia", abbrev: "Saudi Arabia"},
		{pattern: "Scotland", abbrev: "Scot."},
		{pattern: "Senegal", abbrev: "Sen."},
		{pattern: "Serbia", abbrev: "Serb."},
		{pattern: "Seychelles", abbrev: "Sey."},
		{pattern: "Sierra Leone", abbrev: "Sierra Leone"},
		{pattern: "Singapore", abbrev: "Sing."},
		{pattern: "Slovakia", abbrev: "Slovk."},
		{pattern: "Slovenia", abbrev: "Slovn."},
		{pattern: "Solomon Islands", abbrev: "Solom. Is."},
		{pattern: "Somalia", abbrev: "Som."},
		{pattern: "South Africa", abbrev: "S. Afr."},
		{pattern: "South America", abbrev: "S. Am."},
		{pattern: "Spain", abbrev: "Spain"},
		{pattern: "Sri Lanka", abbrev: "Sri Lanka"},
		{pattern: "Sudan", abbrev: "Sudan"},
		{pattern: "Suriname", abbrev: "Surin."},
		{pattern: "Swaziland", abbrev: "Swaz."},
		{pattern: "Sweden", abbrev: "Swed."},
		{pattern: "Switzerland", abbrev: "Switz."},
		{pattern: "Syria", abbrev: "Syria"},
		{pattern: "Taiwan", abbrev: "Taiwan"},
		{pattern: "Tajikistan", abbrev: "Taj."},
		{pattern: "Tanzania", abbrev: "Tanz."},
		{pattern: "Thailand", abbrev: "Thai."},
		{pattern: "Timor-Leste (East Timor)", abbrev: "Timor-Leste"},
		{pattern: "Togo", abbrev: "Togo"},
		{pattern: "Tonga", abbrev: "Tonga"},
		{pattern: "Trinidad & Tobago", abbrev: "Trin. & Tobago"},
		{pattern: "Tunisia", abbrev: "Tunis."},
		{pattern: "Turkey", abbrev: "Turk."},
		{pattern: "Turkmenistan", abbrev: "Turkm."},
		{pattern: "Turks & Caicos Islands", abbrev: "Turks & Caicos Is."},
		{pattern: "Tuvalu", abbrev: "Tuvalu"},
		{pattern: "Uganda", abbrev: "Uganda"},
		{pattern: "Ukraine", abbrev: "Ukr."},
		{pattern: "United Arab Emirates", abbrev: "U.A.E."},
		{pattern: "United Kingdom", abbrev: "U.K."},
		{pattern: "United States of America", abbrev: "U.S."},
		{pattern: "Uruguay", abbrev: "Uru."},
		{pattern: "Uzbekistan", abbrev: "Uzb."},
		{pattern: "Vanuatu", abbrev: "Vanuatu"},
		{pattern: "Vatican City", abbrev: "Vatican"},
		{pattern: "Venezuela", abbrev: "Venez."},
		{pattern: "Vietnam", abbrev: "Viet."},
		{pattern: "Virgin Islands, British", abbrev: "Virgin Is."},
		{pattern: "Wales", abbrev: "Wales"},
		{pattern: "Yemen", abbrev: "Yemen"},
		{pattern: "Zambia", abbrev: "Zam."},
		{pattern: "Zimbabwe", abbrev: "Zim."},
	}

	for i := range lbbCountriesAndRegions {
		entry := &lbbCountriesAndRegions[i]

		pattern := fmt.Sprintf(`(?i)\b%s\b`, entry.pattern)
		entry.re = regexp.MustCompile(pattern)
	}
}
