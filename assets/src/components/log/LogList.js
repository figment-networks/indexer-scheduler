import React from 'react'
import Table from '../table/Table.js'
import PropTypes from 'prop-types'

const logColumns = [{
  headerStyle: { width: '120px' },
  dataField: 'time',
  sort: true,
  text: 'Time'
}, {
  headerStyle: { width: '80px' },
  dataField: 'height',
  sort: true,
  text: 'Height'
}, {
  headerStyle: { width: '50px' },
  dataField: 'retry_count',
  sort: true,
  text: 'Retry'
}, {
  headerStyle: { width: '150px' },
  dataField: 'hash',
  sort: true,
  text: 'Hash'
}, {
  headerStyle: { width: '50px' },
  dataField: 'nonce',
  sort: true,
  text: 'Nonce'
}, {
  headerStyle: { width: '200px' },
  dataField: 'error',
  sort: true,
  style: { width: '30%' },
  text: 'Error'
}]

class LogList extends React.Component {
  static propTypes = {
    logs: PropTypes.array.isRequired
  }

  formatTime (logs) {
    return logs.map(log => {
      log.time = new Date(log.time).toISOString()
      return log
    })
  }

  render () {
    const data = this.formatTime(this.props.logs)
    return (
      <Table columns={logColumns} data={data} tableName="time"/>
    )
  }
}

export default LogList
