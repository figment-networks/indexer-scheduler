import React from 'react'
import Tasks from '../containers/Tasks'
import LastData from '../containers/LastData'
import NewTask from '../containers/NewTask'

import Container from 'react-bootstrap/Container'
import Jumbotron from 'react-bootstrap/Jumbotron'
import Row from 'react-bootstrap/Row'
import Tabs from 'react-bootstrap/Tabs'
import Tab from 'react-bootstrap/Tab'

const App = () => (
  <Container>
    <Jumbotron>
        <h1>Scheduler UI</h1>
    </Jumbotron>
   
    <Row>
        <Tasks />
    </Row>
    
    <Row>
        <LastData />
    </Row>



  </Container>
)

export default App
