import dayjs, { type Dayjs } from "dayjs";
import * as d3 from "d3";
import { Delaunay } from "d3";
import _ from "lodash";
import tippy from "tippy.js";
import COLORS from "./colors";
import {
  formatCurrency,
  formatCurrencyCrude,
  isMobile,
  now,
  svgUrl,
  tooltip,
  type Legend,
  type Networth
} from "./utils";

export interface NetworthChartCallbacks {
  onHover?: (point: Networth, series: "networth" | "investment") => void;
  onLeave?: () => void;
  onRangeSelect?: (from: Dayjs, to: Dayjs) => void;
}

export function selectionToInclusiveDateRange(
  selection: [number, number],
  scale: d3.ScaleTime<number, number>,
  dates: Dayjs[]
): { from: Dayjs; to: Dayjs } | null {
  if (dates.length === 0) return null;

  const [fromX, toX] = selection[0] <= selection[1] ? selection : [selection[1], selection[0]];
  const fromDate = dayjs(scale.invert(fromX));
  const toDate = dayjs(scale.invert(toX));
  const from = dates.reduce((closest, date) =>
    Math.abs(date.diff(fromDate)) < Math.abs(closest.diff(fromDate)) ? date : closest
  );
  const to = dates.reduce((closest, date) =>
    Math.abs(date.diff(toDate)) < Math.abs(closest.diff(toDate)) ? date : closest
  );

  return from.isAfter(to, "day") ? { from: to, to: from } : { from, to };
}

function networth(d: Networth) {
  return d.investmentAmount + d.gainAmount - d.withdrawalAmount;
}

function investment(d: Networth) {
  return d.investmentAmount - d.withdrawalAmount;
}

export function renderNetworth(
  points: Networth[],
  element: Element,
  callbacks: NetworthChartCallbacks = {}
): { destroy: () => void; legends: Legend[] } {
  const start = _.min(_.map(points, (p) => p.date));
  const last = _.max(_.map(points, (p) => p.date));
  const end = last?.isSame(start) ? last.add(1, "day") : last || now();

  const svg = d3.select(element);

  svg.selectAll("*").remove();

  const right = isMobile() ? 10 : 80,
    margin = { top: 15, right: right, bottom: 20, left: 40 },
    width = Math.max(element.parentElement.clientWidth, 800) - margin.left - margin.right,
    height = +svg.attr("height") - margin.top - margin.bottom,
    g = svg.append("g").attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  svg.attr("width", width + margin.left + margin.right);

  const areaKeys = ["gain", "loss"];
  const colors = [COLORS.gain, COLORS.loss];
  const areaScale = d3.scaleOrdinal<string>().domain(areaKeys).range(colors);

  const lineKeys = ["networth", "investment"];
  const lineScale = d3
    .scaleOrdinal<string>()
    .domain(lineKeys)
    .range([COLORS.primary, COLORS.secondary]);

  const positions = _.flatMap(points, (p) => [
    p.gainAmount + p.investmentAmount - p.withdrawalAmount,
    p.investmentAmount - p.withdrawalAmount
  ]);
  positions.push(0);

  const x = d3.scaleTime().range([0, width]).domain([start, end]),
    y = d3.scaleLinear().range([height, 0]).domain(d3.extent(positions)),
    z = d3.scaleOrdinal<string>(colors).domain(areaKeys);

  const area = (y0: number, y1: (d: Networth) => number) =>
    d3
      .area<Networth>()
      .curve(d3.curveMonotoneX)
      .x((d) => x(d.date))
      .y0(y0)
      .y1(y1);

  g.append("g")
    .attr("class", "axis x")
    .attr("transform", "translate(0," + height + ")")
    .call(d3.axisBottom(x));

  if (!isMobile()) {
    g.append("g")
      .attr("class", "axis y")
      .attr("transform", `translate(${width},0)`)
      .call(d3.axisRight(y).tickPadding(5).tickFormat(formatCurrencyCrude));
  }

  g.append("g")
    .attr("class", "axis y")
    .call(d3.axisLeft(y).tickSize(-width).tickFormat(formatCurrencyCrude));

  const layer = g.selectAll(".layer").data([points]).enter().append("g").attr("class", "layer");

  const clipAboveID = _.uniqueId("clip-above");
  layer
    .append("clipPath")
    .attr("id", clipAboveID)
    .append("path")
    .attr(
      "d",
      area(height, (d) => {
        return y(d.gainAmount + d.investmentAmount - d.withdrawalAmount);
      })
    );

  const clipBelowID = _.uniqueId("clip-below");
  layer
    .append("clipPath")
    .attr("id", clipBelowID)
    .append("path")
    .attr(
      "d",
      area(0, (d) => {
        return y(d.gainAmount + d.investmentAmount - d.withdrawalAmount);
      })
    );

  layer
    .append("path")
    .style("fill", z("gain"))
    .style("opacity", "0.2")
    .attr("clip-path", svgUrl(clipAboveID))
    .attr(
      "d",
      area(0, (d) => {
        return y(d.investmentAmount - d.withdrawalAmount);
      })
    );

  layer
    .append("path")
    .attr("clip-path", svgUrl(clipBelowID))
    .style("fill", z("loss"))
    .style("opacity", "0.2")
    .attr(
      "d",
      area(height, (d) => {
        return y(d.investmentAmount - d.withdrawalAmount);
      })
    );

  layer
    .append("path")
    .style("stroke", lineScale("investment"))
    .style("stroke-width", "1.5")
    .style("fill", "none")
    .attr(
      "d",
      d3
        .line<Networth>()
        .curve(d3.curveMonotoneX)
        .x((d) => x(d.date))
        .y((d) => y(investment(d)))
    );

  layer
    .append("path")
    .style("stroke", lineScale("networth"))
    .style("stroke-width", "1.5")
    .style("fill", "none")
    .attr(
      "d",
      d3
        .line<Networth>()
        .curve(d3.curveMonotoneX)
        .x((d) => x(d.date))
        .y((d) => y(networth(d)))
    );

  const hoverCircle = layer
    .append("circle")
    .attr("r", "3")
    .attr("fill", "none")
    .style("pointer-events", "none");
  const selectionOverlay = g
    .append("rect")
    .attr("class", "networth-selection")
    .attr("y", 0)
    .attr("height", height)
    .style("display", "none");
  const chartDates = points.map((point) => point.date);
  const chartNode = svg.node() as SVGSVGElement;
  let dragStartX: number | null = null;
  let dragging = false;
  const dragThreshold = 8;

  const chartX = (event: PointerEvent) => {
    const bounds = chartNode.getBoundingClientRect();
    return Math.max(
      0,
      Math.min(
        width,
        ((event.clientX - bounds.left) / bounds.width) * (width + margin.left + margin.right) -
          margin.left
      )
    );
  };
  const finishDrag = (event: PointerEvent) => {
    if (dragStartX === null) return;
    const endX = chartX(event);
    if (dragging) {
      const range = selectionToInclusiveDateRange([dragStartX, endX], x, chartDates);
      if (range) callbacks.onRangeSelect?.(range.from, range.to);
    }
    selectionOverlay.style("display", "none");
    dragStartX = null;
    dragging = false;
  };
  const onPointerDown = (event: PointerEvent) => {
    if (event.button !== 0) return;
    dragStartX = chartX(event);
    dragging = false;
    chartNode.setPointerCapture?.(event.pointerId);
  };
  const onPointerMove = (event: PointerEvent) => {
    if (dragStartX === null) return;
    const currentX = chartX(event);
    if (!dragging && Math.abs(currentX - dragStartX) < dragThreshold) return;
    dragging = true;
    selectionOverlay
      .style("display", null)
      .attr("x", Math.min(dragStartX, currentX))
      .attr("width", Math.abs(currentX - dragStartX));
  };
  chartNode.addEventListener("pointerdown", onPointerDown);
  chartNode.addEventListener("pointermove", onPointerMove);
  chartNode.addEventListener("pointerup", finishDrag);
  chartNode.addEventListener("pointercancel", finishDrag);
  let pointerX = 0;
  let pointerY = 0;
  const pointerRect = () => new DOMRect(pointerX, pointerY, 0, 0);
  const t = tippy(hoverCircle.node(), {
    theme: "light",
    delay: 0,
    allowHTML: true,
    getReferenceClientRect: pointerRect
  });

  const networthVoronoiPoints: Delaunay.Point[] = _.map(points, (d) => [x(d.date), y(networth(d))]);
  const investmentVoronoiPoints: Delaunay.Point[] = _.map(points, (d) => [
    x(d.date),
    y(investment(d))
  ]);
  const voronoi = Delaunay.from(networthVoronoiPoints.concat(investmentVoronoiPoints)).voronoi([
    0,
    0,
    width,
    height
  ]);

  layer
    .append("g")
    .selectAll("path")
    .data(
      points.map((p) => ["networth", p]).concat(points.map((p) => ["investment", p])) as [
        string,
        Networth
      ][]
    )
    .enter()
    .append("path")
    .style("pointer-events", "all")
    .style("fill", "none")
    .attr("data-series", ([pointType]) => pointType)
    .attr("data-point-index", (_, i) => (i % points.length).toString())
    .attr("d", (_, i) => {
      return voronoi.renderCell(i);
    })
    .on("mousemove", (event: MouseEvent) => {
      pointerX = event.clientX;
      pointerY = event.clientY;
      t.popperInstance?.forceUpdate();
    })
    .on("mouseover", (event: MouseEvent, [pointType, d]) => {
      pointerX = event.clientX;
      pointerY = event.clientY;
      hoverCircle
        .attr("cx", x(d.date))
        .attr("cy", y(pointType == "networth" ? networth(d) : investment(d)))
        .attr("fill", lineScale(pointType));

      t.setProps({
        placement: pointType == "networth" ? "top" : "bottom",
        content: tooltip([
          ["Date", d.date.format("DD MMM YYYY")],
          ["Net Worth", [formatCurrency(networth(d)), "has-text-weight-bold has-text-right"]],
          [
            "Net Investment",
            [formatCurrency(investment(d)), "has-text-weight-bold has-text-right"]
          ],
          ["Gain / Loss", [formatCurrency(d.gainAmount), "has-text-weight-bold has-text-right"]]
        ])
      });
      t.show();
      callbacks.onHover?.(d, pointType as "networth" | "investment");
    })
    .on("mouseout", () => {
      t.hide();
      hoverCircle.attr("fill", "none");
      callbacks.onLeave?.();
    });

  const legends: Legend[] = [
    {
      label: "Net Worth",
      color: lineScale("networth"),
      shape: "line"
    },
    {
      label: "Net Investment",
      color: lineScale("investment"),
      shape: "line"
    },
    {
      label: "Gain",
      color: areaScale("gain"),
      shape: "square"
    },
    {
      label: "Loss",
      color: areaScale("loss"),
      shape: "square"
    }
  ];

  const destroy = () => {
    chartNode.removeEventListener("pointerdown", onPointerDown);
    chartNode.removeEventListener("pointermove", onPointerMove);
    chartNode.removeEventListener("pointerup", finishDrag);
    chartNode.removeEventListener("pointercancel", finishDrag);
    t.destroy();
  };

  return { destroy, legends };
}
