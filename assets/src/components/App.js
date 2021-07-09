import React from 'react'
import Tasks from './task/Tasks'
import Header from './header/Header'

import '../style/style.css'

class App extends React.Component {
  render () {
    return (
      <div className="background-light container">
        <Header/>
        <Tasks />
      </div>
    )
  }
}

export default App
