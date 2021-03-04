import React from 'react'
import PropTypes from 'prop-types'
import { Table } from "react-bootstrap";


class TaskList extends React.Component {

  clickLoadTaskInformation(task_id, network, chain_id, kind, e) {
    this.props.loadTaskInformation(task_id, network, chain_id, kind)
  }

  render() {
    return (
      <Table striped bordered condensed hover>
    <thead>
    <tr>
        <th>id</th>
        <th>network</th>
        <th>chain_id</th>
        <th>kind</th>
        <th>duration</th>
        <th>status</th>
    </tr>
    </thead>
    {this.props.tasks.map((task, i) =>
      <tr key={task.id} onClick={(e) => this.clickLoadTaskInformation(task.task_id, task.network, task.chain_id, task.kind, e)}>
      <td>{task.task_id}</td>
      <td>{task.network}</td>
      <td>{task.chain_id}</td>
      <td>{task.kind}</td>
      <td>{task.duration}</td>
      <td>{task.status} </td>
      </tr>
    )}
  </Table>
    );
  }
}

export default TaskList
