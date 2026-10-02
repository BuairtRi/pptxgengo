package wmdesign

import "encoding/json"

// SceneDataReference is a fictional, bounded inspection packet. It exercises
// source adapters; it is not a template library or native qualification claim.
func SceneDataReference(year int) Document {
	d := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: year}
	node := func(id, body string) Node {
		return Node{ID: id, Kind: "scene", Scene: &SceneSpec{Node: json.RawMessage(body), Path: "/synthetic/data/" + id}}
	}
	slide := func(id, title string, nodes ...Node) {
		d.Slides = append(d.Slides, SlideSpec{ID: "scene-data-" + id, Eyebrow: "WMDS data component inspection", Title: title, Frame: FrameRequest{Rail: "none", Footer: "compact"}, ContentKind: "synthetic_example", Source: "Illustrative scenario; synthetic values and people", Nodes: nodes})
	}
	slide("card-context", "Cards retain their source anatomy",
		node("context-narrow", `{"type":"card","x":57,"y":144,"w":162,"h":216,"surface":"subtle","title":"Diagnose","inlineNumber":"01","body":[{"p":"Name the delay and agree the next review."}]}`),
		node("title-only", `{"type":"card","x":237,"y":144,"w":198,"h":216,"surface":"inverse","band":{"surface":"deep"},"label":"Workstream","title":"Shared view"}`),
		node("featured", `{"type":"card","x":453,"y":144,"w":216,"h":216,"surface":"subtle","state":"featured","tag":"Selected","title":"Pilot","body":[{"checklist":[{"text":"Owner named","on":true},{"text":"Review booked","on":true},{"text":"Data ready","on":false}]}]}`),
		node("placeholder", `{"type":"card","x":687,"y":144,"w":216,"h":216,"surface":"light","state":"placeholder","title":"Next wave","placeholder":"To be confirmed"}`))
	slide("card-compound", "Compound card copy stays editable",
		node("body-columns", `{"type":"card","x":57,"y":144,"w":270,"h":306,"surface":"outline","label":"Two views","title":"Shared plan","body":[{"columns":[[{"label":"Now"},{"p":"Agree scope and name owners."}],[{"label":"Next"},{"p":"Run a pilot and review results."}]]}]}`),
		node("quote", `{"type":"card","x":345,"y":144,"w":270,"h":306,"surface":"inverse","quote":{"text":"Make every decision visible.","by":"Illustrative program lead"}}`),
		node("case", `{"type":"card","x":633,"y":144,"w":270,"h":306,"surface":"subtle","case":{"client":"Illustrative team","challenge":"Reviews started too late.","did":["Moved reviews earlier","Named exception owners"],"results":[{"value":"4","label":"days saved"},{"value":"2","label":"pilot teams"}]}}`))
	slide("card-metrics", "Numbers use authored card geometry",
		node("metric-card", `{"type":"card","x":57,"y":144,"w":270,"h":270,"surface":"inverse","metric":{"value":"64%","label":"manual steps removed","secondary":{"value":"6 mo","label":"to payback"},"status":"on"}}`),
		node("metric-group", `{"type":"card","x":345,"y":144,"w":558,"h":270,"surface":"subtle","title":"Capacity returned","body":[{"p":"The scenario separates time returned from cash savings."}],"metricGroup":{"primary":{"value":"$2.4M","label":"value potential"},"secondary":[{"value":"18%","label":"less effort"},{"value":"6 mo","label":"to payback"}]}}`))
	slide("native-fees", "Fee tables retain native cells",
		node("fee-table", `{"type":"table","x":57,"y":144,"w":558,"header":"dark","preset":"fees","cols":[{"k":"phase","label":"Phase","w":270},{"k":"share","label":"% of fee","w":108,"type":"num"},{"k":"amount","label":"Amount","w":180,"type":"num"}],"rows":[{"phase":"Discovery","share":"25%","amount":"$60K"},{"phase":"Pilot","share":"50%","amount":"$120K"},{"phase":"Transition","share":"25%","amount":"$60K"},{"phase":"Illustrative total","share":"100%","amount":"$240K","total":true}]}`),
		node("fee-summary", `{"type":"feesummary","x":633,"y":144,"w":270,"model":"Illustrative fixed fee","total":"$240K","totalLabel":"for three phases","lines":[["Duration","8 weeks"],["Team","4 roles"]],"terms":["Milestones agreed","Scope changes reviewed"]}`))
	slide("native-indicators", "Native table cells and indicators",
		node("indicator-table", `{"type":"table","x":57,"y":144,"w":846,"header":"dark","rowH":54,"rowHeader":true,"cols":[{"k":"work","label":"Work","w":126},{"k":"status","label":"Status","w":126,"type":"status"},{"k":"allocation","label":"Allocation","w":126,"type":"allocation"},{"k":"maturity","label":"Maturity","w":162,"type":"maturity"},{"k":"rating","label":"Rating","w":108,"type":"rating"},{"k":"check","label":"Ready","w":90,"type":"check"},{"k":"coverage","label":"Coverage","w":108,"type":"harvey"}],"rows":[{"work":"Pilot A","status":"on","allocation":0.75,"maturity":[2,4],"rating":3,"check":true,"coverage":3},{"work":"Pilot B","status":"risk","allocation":0.5,"maturity":[1,3],"rating":2,"check":false,"coverage":2},{"work":"Pilot C","status":"off","allocation":0.25,"maturity":[0,2],"rating":1,"check":false,"coverage":1}]}`))
	slide("native-table-bullets", "Merged headings and native bullets",
		node("responsibility-table", `{"type":"table","x":57,"y":144,"w":846,"header":"light","dense":true,"rowH":72,"rowHeader":true,"groups":[{"label":"Illustrative responsibilities","from":1,"to":2}],"cols":[{"k":"role","label":"Role","w":198},{"k":"now","label":"Pilot","w":324,"type":"bullets"},{"k":"next","label":"Transition","w":324,"type":"bullets"}],"rows":[{"role":"Program lead","now":["Agree scope","Name decision owners"],"next":["Review results","Confirm next wave"]},{"role":"Delivery team","now":["Create shared views","Track open questions"],"next":["Transfer operating knowledge","Document review routines"]}]}`))
	slide("native-column", "Column data remains in a workbook",
		node("clustered-chart", `{"type":"chart","kind":"column","mode":"clustered","x":57,"y":144,"w":846,"h":288,"title":"Illustrative cycle times","units":"Days","categories":["Wave 1","Wave 2","Wave 3"],"series":[{"name":"Before","values":[12,11,14]},{"name":"Pilot","values":[5,6,7]}],"colors":["deemph.3","series.1"]}`))
	slide("native-line", "Line series retain native dash styles",
		node("line-chart", `{"type":"chart","kind":"line","x":57,"y":144,"w":846,"h":270,"title":"Illustrative effort transition","units":"Relative effort","categories":["Wk 1","Wk 4","Wk 8","Wk 12"],"series":[{"name":"Delivery","values":[2,6,5,2],"color":"series.1"},{"name":"Operations","values":[1,3,5,6],"color":"deemph.3","dashed":true}],"yMin":0}`))
	slide("qualitative-quadrant", "Qualitative positions stay editable",
		node("quadrant-chart", `{"type":"chart","kind":"quadrant","x":57,"y":144,"w":846,"h":288,"title":"Illustrative priorities","xTitle":"Effort","yTitle":"Value","key":true,"quadrants":{"tl":{"name":"Quick wins","state":"desirable","tag":"Aim here"},"tr":{"name":"Big bets"},"bl":{"name":"Fill-ins"},"br":{"name":"Avoid","state":"undesirable","tag":"Review"}},"items":[{"x":0.2,"y":0.8,"label":"Shared calendar"},{"x":0.7,"y":0.7,"label":"Workflow redesign"},{"x":0.3,"y":0.3,"label":"Field clean-up"}]}`))
	return d
}
