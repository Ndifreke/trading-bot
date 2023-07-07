import * as d3 from "d3";
import { useEffect, useLayoutEffect, useRef } from "react";

const data = [1, 2, 3, 4, 5];

const D3 = () => {
  const me = useRef<any>();

  useLayoutEffect(() => {

  }, []);

  return (
    <div>
      <svg ref={me} height={300} width={300} style={{ border: "1px solid red" }} viewBox="-35 -30 300 300">
        {/* <g>
          <rect height={50} width={200} fill="red" />
          <rect height={50} width={200} fill="blue" />
          <circle cx={40} cy={50} r={25} fill="black" />
          <circle cx={200 - 40} cy={50} r={25} fill="black" />
          <text y={25} x={200}>
            Hello Friend
          </text>
        </g> */}
      </svg>
    </div>
  );
};

export default D3;
