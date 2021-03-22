import React from 'react'
import PropTypes from 'prop-types'
import { Table } from "react-bootstrap";
import Button from 'react-bootstrap/Button'


class TaskList extends React.Component {

  clickLoadTaskInformation(task_id, network, chain_id, kind, e) {
    this.props.loadTaskInformation(task_id, network, chain_id, kind)
  }
  clickEnableTask(task_id, e) {
    this.props.enableTask(task_id)
  }
  clickDisableTask(task_id, e) {
    this.props.disableTask(task_id)
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
        <th>enabled</th>
        <th>config</th>
        <th></th>
    </tr>
    </thead>
    {this.props.tasks.map((task, i) =>
      <tr key={i}>
        <td>{task.task_id}</td>
        <td>{task.network}</td>
        <td>{task.chain_id}</td>
        <td>{task.kind}</td>
        <td>{task.duration}</td>
        <td>{task.status} </td>
        <td>
          {task.enabled
            ? <Button onClick={(e) => this.clickDisableTask(task.id, e)} >enabled</Button>
            : <Button onClick={(e) => this.clickEnableTask(task.id, e)} >disabled</Button>
          } </td>
        <td>{JSON.stringify(task.config)}</td>
        <td ><Button onClick={(e) => this.clickLoadTaskInformation(task.task_id, task.network, task.chain_id, task.kind, e)}  >See </Button> </td>
      </tr>
    )}
  </Table>
    );
  }
}

export default TaskList
