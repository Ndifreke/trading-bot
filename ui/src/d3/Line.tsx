import * as d3 from "d3";
import { useEffect, useLayoutEffect, useMemo } from "react";

const dms = {
  width: 600,
  height: 400,
  marginTop: 20,
  marginRight: 20,
  marginBottom: 50,
  marginLeft: 50,
};

// xRange = [marginLeft, width - marginRight], // [left, right] shift by ml to w - mr
// yType = d3.scaleLinear, // the y-scale type
// yDomain, // [ymin, ymax]
// yRange = [height - marginBottom, marginTop], // [bottom, top]

const data = [0,0.5,0.5,1, 1.3, 1.6, null, 1.3, 0.6, 2,3,4,5,6 ] as unknown as number[];
const boundedWidth = dms.width - (dms.marginLeft - dms.marginRight);
const boundedHeight = dms.height - (dms.marginTop - dms.marginBottom);

const move = (x: number, y: number) => {
  return `translate(${x}, ${y})`;
};
const Lines = () => {
  useLayoutEffect(() => {
    const svg = d3.select("#chart");
    svg.selectAll("*").remove();
    svg.attr("viewbox", [0, 0, dms.width, dms.height]).attr("width", dms.width).attr("height", 500);
    const xScale = d3.scaleLinear().domain([0, 5]).range([0, boundedWidth])
   

    const yScale = d3.scaleLinear().domain([0, 5]).range([boundedHeight, 0]);

    const axisBottom = d3.axisBottom(xScale).ticks(5);
    const xAxisTop = d3.axisLeft(yScale).ticks(5);
    svg.append("g").attr("id", "x-axis").call(axisBottom).attr("transform", move(dms.marginLeft, 480));
    svg.append("g").attr("id", "y-axis").call(xAxisTop).attr("transform", move(50, 50));

    svg
      .append("g")
      .selectAll("circle")
      .data(data)
      .enter()
      .append("circle")
      .attr("cx", (d) => xScale(d))
      .attr("cy", (d) => yScale(d))
      .attr("r", 9)
      .attr("fill", "darkred");

      //@ts-ignore
      const line =  d3.line(d => xScale(d), d => yScale(d))

    svg
      .append("path")
      //@ts-ignore
      .attr("d", `${line(data)}`)
      .attr("stroke", "black");
  }, []);

  return (
    <div
    // style={{margin:"20px"}}
    >
      <svg id="chart"></svg>{" "}
    </div>
  );
};

export default Lines;
