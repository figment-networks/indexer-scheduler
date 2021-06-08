import React from 'react'
import PropTypes from 'prop-types'
import { Table } from "react-bootstrap";

const LastdataList = ({lastdatas}) => (
  <Table striped bordered condensed hover>
    <thead>
    <tr>
        <th>time</th>
        <th>retry</th>
        <th>height</th>
        <th>hash</th>
        <th>error</th>
        <th>nonce</th>
    </tr>
    </thead>
    {lastdatas.map((ld, i) =>
      <tr key={i}>
      <td>{ld.time}</td>
      <td>{ld.retry_count}</td>
      <td>{ld.height}</td>
      <td>{ld.hash}</td>
      <td>{ld.error}</td>
      <td>{ld.none}</td>
      </tr>
    )}
  </Table>
)

LastdataList.propTypes = {
    tasks: PropTypes.array.isRequired
}

export default LastdataList
