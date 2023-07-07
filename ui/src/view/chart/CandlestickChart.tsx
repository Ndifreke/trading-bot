//@ts-nocheck
import React from "react";
import { ChartCanvas, Chart } from "react-stockcharts";
import { CandlestickSeries, LineSeries } from "react-stockcharts/lib/series";
import { XAxis, YAxis } from "react-stockcharts/lib/axes";
import { discontinuousTimeScaleProvider } from "react-stockcharts/lib/scale";
import { fitWidth } from "react-stockcharts/lib/helper";
import { last } from "react-stockcharts/lib/utils";
import { CrossHairCursor, EdgeIndicator, CurrentCoordinate, MouseCoordinateX, MouseCoordinateY } from "react-stockcharts/lib/coordinates";
import { LabelAnnotation, Label, Annotate } from "react-stockcharts/lib/annotation";
import { ema, heikinAshi, sma } from "react-stockcharts/lib/indicator";
import * as d3 from "d3";

const Median = LineSeries;

type Kline = {
  openTime: number;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
  closeTime: number;
  quoteAssetVolume: number;
  tradeNum: number;
  takerBuyBaseAssetVolume: number;
  takerBuyQuoteAssetVolume: number;
};
type GraphType = {
  data: Kline[];
  entryPoints: { dipLowerPrice: number; dipLowPrice: number; gainLowPrice: number; gainHighPrice: number };
  trendName: string;
  priceMidPoint: number;
};

type CandlestickChartProps = {
  graph: GraphType;

  width: number;
  height: number;
  ratio: number;
};

const timeFormats = {
  tooltip: d3.timeFormat("%H:%M:%S"),
  tooltipLong: d3.timeFormat("%H:%M %b %d %Y"),
  seconds: d3.timeFormat("%H:%M"),
  minutes: d3.timeFormat("%H:%M"),
  hours: d3.timeFormat("%H:%M"),
  days: d3.timeFormat("%b %d"),
  weeks: d3.timeFormat("%b %d"),
  months: d3.timeFormat("%b %d"),
};
const CandlestickChart: React.FC<CandlestickChartProps> = ({ graph, width, height, ratio }) => {
  const xScaleProvider = discontinuousTimeScaleProvider.inputDateAccessor((d: Kline) => new Date(d.openTime));

  const { data, xScale, xAccessor, displayXAccessor } = xScaleProvider(graph.data);
  const { entryPoints } = graph;
  const xExtents = [xAccessor(last(data)), xAccessor(data[Math.max(0, data.length - 100)])];
  console.log(entryPoints.gainHighPrice);

 
  return (
    <ChartCanvas
      height={height}
      width={width}
      ratio={ratio}
      margin={{ left: 150, right: 150, top: 150, bottom: 150 }}
      seriesName="MSFT"
      data={data}
      xScale={xScale}
      xAccessor={xAccessor}
      displayXAccessor={displayXAccessor}
      xExtents={xExtents}
    >
      <Chart id={1} yExtents={(d: Kline) => [d.high, d.low]}>
        <EdgeIndicator
          itemType="first"
          orient="right"
          edgeAt="right"
          yAccessor={(d) => entryPoints.gainHighPrice}
          displayFormat={() => `H-G-E ${entryPoints.gainHighPrice}`}
          fill={(d) => "#6BA583"}
        />
{/* 
        <EdgeIndicator
          itemType="first"
          orient="left"
          edgeAt="left"
          yAccessor={(d) => entryPoints.gainLowPrice}
          displayFormat={() => `L-G-E ${entryPoints.gainLowPrice}`}
          fill={(d) => "#6BA583"}
        /> */}

        <EdgeIndicator
          itemType="first"
          orient="right"
          edgeAt="right"
          yAccessor={(d) => entryPoints.dipLowPrice}
          displayFormat={() => `L-D-E ${entryPoints.dipLowPrice}`}
          fill={(d) => "#FF0000"}
        />

        {/* <EdgeIndicator
          itemType="first"
          orient="left"
          edgeAt="left"
          yAccessor={(d) => entryPoints.dipLowerPrice}
          displayFormat={() => `LW-D-E ${entryPoints.dipLowerPrice}`}
          fill={(d) => "#FF0000"}
        />  */}
        <XAxis

          axisAt="bottom"
          orient="bottom"
          ticks={6}
        />
        <YAxis axisAt="left" orient="left" ticks={6} />

        <CandlestickSeries />
        <Median yAccessor={() => graph.priceMidpoint} fill={(d) => "orange"} />
      </Chart>
    </ChartCanvas>
  );
};

export default fitWidth(CandlestickChart);
