import { Box, Text } from '@chakra-ui/react';
import { useMemo } from 'react';
import graph from "../dump/kline.json";
import CandlestickChart from './chart/CandlestickChart';

const App = () => {
const {height, width} = useMemo(()=> ({height:window.innerHeight, width:window.innerWidth}),[])


return (
  <Box>
    <Text>{graph.trendName}</Text>
     <CandlestickChart graph={graph} height={height - 100} width={width - 100} ratio={9} />
     </Box>
  );
};

export default App;
