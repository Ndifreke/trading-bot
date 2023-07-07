import {
  ChakraProvider,
  theme
} from "@chakra-ui/react"
import CandlestickChart from "./view/Page"
import D3 from "./d3/D3"
import Lines from "./d3/Line"

export const App = () => (
  <ChakraProvider theme={theme}>
    {/* <CandlestickChart/> */}
    {/* <D3/> */}
    <Lines/>
  </ChakraProvider>
)
