import React from 'react'
import Tasks from '../containers/Tasks'
import LastData from '../containers/LastData'

import Container from 'react-bootstrap/Container'
import Row from 'react-bootstrap/Row'

const App = () => (
  <Container>
    <Row>
      <Tasks />
    </Row>
    <Row>
      <LastData />
    </Row>
  </Container>
)

export default App
