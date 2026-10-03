/**
 * @wmds/slides: West Monroe slides as React components.
 *
 * Every component is a thin wrapper over the reference board's renderer (explorations/components.src.html),
 * so what renders here is exactly what the design system specifies. Props mirror the JSON node spec that
 * pptxgengo turns into native PowerPoint: a slide composed here maps 1:1 onto an editable .pptx slide.
 *
 *   <Slide eyebrow="Our approach" title="Four phases take the close from [[12 days to 5]]" emphasis="highlight">
 *     <Card x={57} y={126} w={270} h={198} surface="subtle" title="Mobilize" body={[{ p: "Team and data in place." }]} />
 *   </Slide>
 *
 * Coordinates are slide points: the slide is 960 x 540 pt and scales to the width of its container.
 */
import * as React from "react";
import { createRenderer } from "./generated/renderer.js";
import { DATA, TEMPLATES, CATALOG } from "./generated/data.js";
import { PHOTOS, ICONS } from "./generated/assets.js";
import { NODE_TYPES } from "./generated/nodes";
import type * as N from "./generated/nodes";
import "./generated/styles.css";

export type * from "./generated/nodes";

let R: any = null;
function renderer() {
  if (!R) R = createRenderer({ tokens: DATA.tokens, frames: DATA.frames, inkRules: DATA.inkRules, icons: ICONS, photos: { ...PHOTOS } });
  return R;
}

type AnyNode = Record<string, unknown> & { type: string };
const NODE_KEY = "__wmNode";

/** Turn JSX children (node components, fragments, arrays) into renderer nodes. */
function collect(children: React.ReactNode): AnyNode[] {
  const out: AnyNode[] = [];
  React.Children.forEach(children, (child) => {
    if (!React.isValidElement(child)) return;
    const t = (child.type as any)?.[NODE_KEY];
    const props = child.props as Record<string, unknown> & { children?: React.ReactNode };
    if (t) {
      const { children: kids, ...rest } = props;
      const n: AnyNode = { type: t, ...rest };
      if (typeof kids === "string" && n.text === undefined) n.text = kids;
      out.push(n);
    } else if (child.type === React.Fragment) {
      out.push(...collect(props.children));
    }
  });
  return out;
}

type StageSpec = { label: string; h: number | "auto"; surface?: string; nodes: AnyNode[] };

/** Renders one renderer stage into a div; re-renders when the spec changes. */
function Stage({ spec, warnings, className, style }: { spec: StageSpec; warnings?: boolean; className?: string; style?: React.CSSProperties }) {
  const ref = React.useRef<HTMLDivElement>(null);
  const key = JSON.stringify(spec);
  React.useLayoutEffect(() => {
    const el = ref.current;
    if (!el) return;
    el.innerHTML = "";
    const spec0 = JSON.parse(key);
    const auto = spec0.h === "auto"; // stage() replaces h with its own estimate, so read it first
    let s = renderer().stage(spec0);
    el.appendChild(s);
    if (auto) {
      // Not every node reports its height to the renderer, so measure what was drawn and redraw at that height.
      const ptpx = s.clientWidth / 960 || 1;
      let bottom = 0;
      Array.prototype.forEach.call(s.children, (c: HTMLElement) => {
        if (c.classList && c.classList.contains("overlay")) return;
        const b = (c.offsetTop || 0) + (c.offsetHeight || 0);
        if (b > bottom) bottom = b;
      });
      const hpt = Math.max(18, Math.ceil((bottom / ptpx + 18) / 18) * 18);
      el.innerHTML = "";
      s = renderer().stage({ ...JSON.parse(key), h: hpt });
      el.appendChild(s);
    }
    if (warnings) requestAnimationFrame(() => renderer().attachWarnings(s));
  }, [key, warnings]);
  return <div ref={ref} className={"wm-root" + (warnings ? " wm-warnings" : "") + (className ? " " + className : "")} style={style} />;
}

export interface SlideProps extends N.SlideProps {
  /** Node components to place on the slide (Card, Table, Text...). They are drawn after any `body` data. */
  children?: React.ReactNode;
  /** Show the renderer's rule warnings (contrast, off-palette, highlight surface) as tags. Off by default. */
  warnings?: boolean;
  className?: string;
  style?: React.CSSProperties;
}

/**
 * One 960 x 540 pt slide in a frame (rail x footer) with the required chrome: legal line, logo, page number
 * and whiteboard grid. Pass eyebrow/title/emphasis for the title zone and place components as children.
 */
export function Slide({ children, warnings, className, style, body, ...props }: SlideProps) {
  const nodes = [...((body as AnyNode[]) || []), ...collect(children)];
  const spec: StageSpec = { label: String(props.title || props.eyebrow || "Slide"), h: 540, nodes: [{ type: "slide", ...props, body: nodes } as AnyNode] };
  return <Stage spec={spec} warnings={warnings} className={className} style={style} />;
}
(Slide as any)[NODE_KEY] = undefined;

export interface CanvasProps {
  /** Height in points; "auto" fits the content. The width is always 960 pt, scaled to the container. */
  height?: number | "auto";
  /** Surface of the whole canvas: light (default), subtle, strong, inverse, deep. */
  surface?: "light" | "subtle" | "strong" | "inverse" | "deep";
  children?: React.ReactNode;
  warnings?: boolean;
  className?: string;
  style?: React.CSSProperties;
}

/** A 960 pt wide drawing area with no slide chrome, for showing components or composing part of a slide. */
export function Canvas({ height = "auto", surface, children, warnings, className, style }: CanvasProps) {
  const spec: StageSpec = { label: "Canvas", h: height, surface, nodes: collect(children) };
  return <Stage spec={spec} warnings={warnings} className={className} style={style} />;
}

function node<P>(type: string, name: string): React.FC<P & { children?: React.ReactNode }> {
  const C = (props: P & { children?: React.ReactNode }) => {
    // Rendered on its own (not inside Slide or Canvas): draw it on an auto-height canvas.
    const { children, ...rest } = props as any;
    const n: AnyNode = { type, ...rest };
    if (typeof children === "string" && n.text === undefined) n.text = children;
    return <Stage spec={{ label: name, h: "auto", nodes: [n] }} />;
  };
  (C as any)[NODE_KEY] = type;
  C.displayName = name;
  return C;
}

/** Text in one type style and ink role. `[[…]]` marks the emphasized phrase; `[^1]` a footnote. */
export const Text = node<N.TextProps>("text", "Text");
/** A rule (hairline to accent weight) in an ink role. */
export const Rule = node<N.RuleProps>("rule", "Rule");
/** A small caps label with a rule running to the right. */
export const GroupLabel = node<N.GroupLabelProps>("grouplabel", "GroupLabel");
/** A heading with a Mono number on the same line. */
export const NumberedHeading = node<N.NumberedHeadingProps>("numhead", "NumberedHeading");
/** A column head: accent rule, label and subhead title. */
export const ColumnHeading = node<N.ColumnHeadingProps>("colhead", "ColumnHeading");
/** Label, title and body as one block. */
export const TextBlock = node<N.TextBlockProps>("textblock", "TextBlock");
/** Square-bullet list; items may have a strong lead-in or sub-items. size "small" for dense zones. */
export const Bullets = node<N.BulletsProps>("bullets", "Bullets");
/** Numbered list. */
export const NumberedList = node<N.NumberedListProps>("ol", "NumberedList");
/** Strong numbered list: Mono numbers with a title and text per item. */
export const StrongNumberList = node<N.StrongNumberListProps>("strongnum", "StrongNumberList");
/** Timed agenda rows: a key (time) column and the item. */
export const Schedule = node<N.ScheduleProps>("schedule", "Schedule");
/** Index rows with numbers, titles and page references (agendas). */
export const IndexList = node<N.IndexListProps>("list", "IndexList");
/** The card system: title, body blocks, bands, badges, edges, metrics, bios, case studies, states. */
export const Card = node<N.CardProps>("card", "Card");
/** A row of cards sharing one style, with inline, band or corner numbering. */
export const CardRow = node<N.CardRowProps>("cardrow", "CardRow");
/** A metric: value, label, change, secondary values, target and status. */
export const Metric = node<N.MetricProps>("metric", "Metric");
/** A call-out: value, text and secondary line. */
export const Callout = node<N.CalloutProps>("callout", "Callout");
/** A pull quote with attribution. */
export const PullQuote = node<N.PullQuoteProps>("pullquote", "PullQuote");
/** A legend, horizontal or vertical. */
export const Legend = node<N.LegendProps>("legend", "Legend");
/** A segmented gauge on palette references. */
export const Gauge = node<N.GaugeProps>("gauge", "Gauge");
/** Indicator primitives: Harvey balls, status, ratings. */
export const Indicators = node<N.IndicatorsProps>("indicators", "Indicators");
/** A row of labelled library icons. */
export const IconRow = node<N.IconRowProps>("icons", "IconRow");
/** A person: photo or initials, name, role and org. */
export const Person = node<N.PersonProps>("person", "Person");
/** A role box with an accent edge (org charts, pods). */
export const RoleBox = node<N.RoleBoxProps>("role", "RoleBox");
/** A horizontal stepper with states. */
export const Stepper = node<N.StepperProps>("stepper", "Stepper");
/** A vertical stepper with titles and text. */
export const VerticalStepper = node<N.VerticalStepperProps>("vstepper", "VerticalStepper");
/** A phase heading: number, title, duration and rule. */
export const PhaseHeading = node<N.PhaseHeadingProps>("phasehead", "PhaseHeading");
/** A time axis with periods, today marker and milestones. */
export const TimeAxis = node<N.TimeAxisProps>("timeaxis", "TimeAxis");
/** A native preset shape: rect, circle, homeplate, chevron, arrow, diamond. */
export const Shape = node<N.ShapeProps>("preset", "Shape");
/** A filled block with centered or left text. */
export const Block = node<N.BlockProps>("block", "Block");
/** A chevron; may carry a number, title and subtitle. */
export const Chevron = node<N.ChevronProps>("chevron", "Chevron");
/** An arrow-shaped block with text. */
export const TextArrow = node<N.TextArrowProps>("textarrow", "TextArrow");
/** A connector line: solid, dashed or dotted, heads at either end, elbow routing, label. */
export const Connector = node<N.ConnectorProps>("connector", "Connector");
/** A diagram node: icon, text and sub-text in a box. */
export const DiagramNode = node<N.DiagramNodeProps>("node", "DiagramNode");
/** A framed region: solid, dashed or filled, with a label. */
export const FrameBox = node<N.FrameBoxProps>("frame", "FrameBox");
/** An architecture container: region, boundary, layer, frame or external. */
export const Container = node<N.ContainerProps>("container", "Container");
/** A labelled layer of cells (architecture). */
export const Layer = node<N.LayerProps>("layer", "Layer");
/** A numbered layer row with label and text. */
export const LayerRow = node<N.LayerRowProps>("layerrow", "LayerRow");
/** A delivery pod: title band and roles. */
export const Pod = node<N.PodProps>("pod", "Pod");
/** A database cylinder. */
export const Cylinder = node<N.CylinderProps>("cylinder", "Cylinder");
/** A device frame (laptop, phone). */
export const Device = node<N.DeviceProps>("device", "Device");
/** A screen placeholder. */
export const Screen = node<N.ScreenProps>("screen", "Screen");
/** An isometric plane with label and text. */
export const Plane = node<N.PlaneProps>("plane", "Plane");
/** Tables: groups, row headers, totals, run-rate, rich cell types; presets raci, fees, requirements, capacity, ratecard. */
export const Table = node<N.TableProps>("table", "Table");
/** Charts: column, bar, line, pie, doughnut, scatter, quadrant. One axis; direct labels or a legend. */
export const Chart = node<N.ChartProps>("chart", "Chart");
/** Gantt: workstreams, bars with soft starts and in-progress hatching, gates, milestones, phases, legend. */
export const Gantt = node<N.GanttProps>("gantt", "Gantt");
/** Roadmap horizons. */
export const Horizons = node<N.HorizonsProps>("horizons", "Horizons");
/** Phase columns: objective, activities, deliverables and gates. */
export const Phases = node<N.PhasesProps>("phases", "Phases");
/** Org chart from a tree of roles. */
export const OrgChart = node<N.OrgChartProps>("orgchart", "OrgChart");
/** Governance tiers with cadence, members and decision rights. */
export const GovernanceStack = node<N.GovernanceStackProps>("governance", "GovernanceStack");
/** Fee summary by milestone, phase or aggregate (never per person on fixed fee). */
export const FeeSummary = node<N.FeeSummaryProps>("feesummary", "FeeSummary");
/** Swimlane process: lanes, steps, decisions, pain points and links. */
export const Swimlane = node<N.SwimlaneProps>("swimlane", "Swimlane");
/** Cycle or hub-and-spoke. */
export const Cycle = node<N.CycleProps>("cycle", "Cycle");
/** Pyramid or funnel with descriptions. */
export const Pyramid = node<N.PyramidProps>("pyramid", "Pyramid");
/** Paired before-and-after rows with a transition arrow. */
export const BeforeAfter = node<N.BeforeAfterProps>("beforeafter", "BeforeAfter");
/** A call-out attached to a target by a hand-drawn arrow. */
export const Annotation = node<N.AnnotationProps>("annotation", "Annotation");
/** Rows of small labelled cells: a layer map or a state grid, with an optional outlined row. */
export const Matrix = node<N.MatrixProps>("matrix", "Matrix");
/** A deliverable page thumbnail with caption (image or placeholder kind), optionally stacked. */
export const Thumbnail = node<N.ThumbnailProps>("thumbnail", "Thumbnail");
/** A hand-drawn mark: underscore, circle, spark or an arrow. One mark per slide. */
export const Mark = node<N.MarkProps>("mark", "Mark");
/** A whiteboard dot field. (Slides draw their own; use this only on canvases.) */
export const Whiteboard = node<N.WhiteboardProps>("whiteboard", "Whiteboard");
/** A square: a photo, a surface, or a stat with label. */
export const Square = node<N.SquareProps>("square", "Square");
/** A photo in a frame with a focal point; optional greyscale. */
export const ImageFrame = node<N.ImageFrameProps>("imageframe", "ImageFrame");
/** The West Monroe logo, positive or reverse. */
export const Logo = node<N.LogoProps>("logo", "Logo");
/** Brand artwork such as the reverse tagline. */
export const Art = node<N.ArtProps>("art", "Art");
/** The collaboration review note: a status tab that peeks onto the slide from the pasteboard. */
export const ReviewNote = node<N.ReviewNoteProps>("reviewnote", "ReviewNote");

export interface TemplateProps {
  /** Template key, `id/variant`, e.g. "cover/photo" or "phase-detail/rail-table". See `TEMPLATES`. */
  name: string;
  /** Shallow overrides for the slide's frame and title fields (title, eyebrow, page, emphasis...). */
  overrides?: Partial<N.SlideProps>;
  /** Extra nodes placed after the template's body. */
  children?: React.ReactNode;
  warnings?: boolean;
  className?: string;
  style?: React.CSSProperties;
}

/** Renders a library template by key. Start from the closest template, then compose your own Slide from its parts. */
export function Template({ name, overrides, children, warnings, className, style }: TemplateProps) {
  const [id, variant] = name.split("/");
  const t: any = (TEMPLATES as any[]).find((x) => x.id === id && (!variant || x.variant === variant));
  if (!t) return <div className="wm-root">Unknown template {name}</div>;
  const slide = { ...t.slide, ...(overrides || {}), body: [...(t.slide.body || []), ...collect(children)] };
  return <Stage spec={{ label: t.name, h: 540, nodes: [slide] }} warnings={warnings} className={className} style={style} />;
}

/** Library icon by name (222 brand icons). Grounded on light, White on dark, Magenta for one highlighted icon. */
export function Icon({ name, size = 36, color = "#070154" }: { name: string; size?: number; color?: "#070154" | "#F900D3" | "#FFFFFF" }) {
  const ref = React.useRef<HTMLSpanElement>(null);
  React.useLayoutEffect(() => {
    const el = ref.current;
    if (!el) return;
    el.innerHTML = "";
    const s = renderer().icon(name, size, color);
    s.style.width = size + "px";
    s.style.height = size + "px";
    el.appendChild(s);
  }, [name, size, color]);
  return <span ref={ref} className="wm-root" style={{ display: "inline-block", width: size, height: size }} />;
}

/** Every library template (id, variant, name, tier, purpose, slots, budget and the slide spec). */
export { TEMPLATES, CATALOG, NODE_TYPES };
/** Icon names in the brand library. */
export const ICON_NAMES: string[] = Object.keys(ICONS);
/** Brand photo names usable as `photo` on Square, ImageFrame, Person and card bios. */
export const PHOTO_NAMES: string[] = Object.keys(PHOTOS);
