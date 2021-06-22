import React, { Component } from 'react'
import PropTypes from 'prop-types'
import { connect } from 'react-redux'
import NewTask from '../containers/NewTask'
import TaskList from '../components/TaskList'

import Button from 'react-bootstrap/Button'
import Container from 'react-bootstrap/Container'
import Row from 'react-bootstrap/Row'
import Col from 'react-bootstrap/Col'


import { fetchTasksIfNeeded, invalidateTasks, enableTask, disableTask, showNewTask, hideNewTask  } from '../actions'
import { fetchLastData, invalidateLastdata  } from '../actions/lastdata'


class Tasks extends Component {
  static propTypes = {
    list: PropTypes.array.isRequired,
    isFetching: PropTypes.bool.isRequired,
    dispatch: PropTypes.func.isRequired,
    showNewTaskForm: PropTypes.bool.isRequired
  }

  componentDidMount() {
    this.props.dispatch(fetchTasksIfNeeded())
  }

  componentDidUpdate(prevProps) {
    if (prevProps.list !== this.props.list) {
      this.props.dispatch(fetchTasksIfNeeded())
    }
  }

  handleRefreshClick = e => {
    e.preventDefault()

    const { dispatch } = this.props
    dispatch(invalidateTasks())
    dispatch(fetchTasksIfNeeded())
  }

  handleNewTaskClick = e => {
    this.props.dispatch(showNewTask())
  }

  handleHideTaskClick = e => {
    this.props.dispatch(hideNewTask())
  }

  loadTaskInformation(task_id,  network,  chain_id, kind) {
    const { dispatch } = this.props
    dispatch(invalidateLastdata())
    dispatch(fetchLastData(task_id,  network,  chain_id, kind, 100, 0))
  }

  enableTask(task_id,  network,  chain_id, kind) {
    const { dispatch } = this.props
    dispatch(enableTask(task_id,  network, chain_id, kind))
    dispatch(invalidateTasks())
    dispatch(fetchTasksIfNeeded())
  }

  disableTask(task_id,  network,  chain_id, kind) {
    const { dispatch } = this.props
    dispatch(disableTask(task_id, network, chain_id, kind))
    dispatch(invalidateTasks())
    dispatch(fetchTasksIfNeeded())
  }

  render() {
    const { list, isFetching, showNewTaskForm } = this.props
    const isEmpty =  (list === null || list.length === 0)
    return (
      <Container>
        {showNewTaskForm ? 
          <Row>
            <Col><h2>Create a new task</h2></Col>
            <Col xs={1}><Button style={{padding: 10}} variant="danger" onClick={this.handleHideTaskClick}>Cancel</Button></Col>
            <NewTask/> 
          </Row>
        : 
          <div> 
            <div className="row" >
              <div className="col"><h2>Task list</h2></div>
              <div className="col" style={{textAlign: "right", maxWidth: "120px"}}><Button style={{padding: 10}} variant="outline-dark" onClick={this.handleRefreshClick}>Refresh</Button></div>
              <div className="col" style={{textAlign: "right", maxWidth: "120px"}}><Button style={{padding: 10}} variant="success" onClick={this.handleNewTaskClick}>New task</Button></div>
            </div>
            <Row>
            {isEmpty
              ? (isFetching ? <h2>Loading...</h2> : "")
              : <TaskList
                  tasks={list}
                  loadTaskInformation={(task_id,  network,  chain_id, kind) => this.loadTaskInformation(task_id,  network,  chain_id, kind)}
                  enableTask={(task_id,  network,  chain_id, kind) => this.enableTask(task_id,  network,  chain_id, kind)}
                  disableTask={(task_id,  network,  chain_id, kind) => this.disableTask(task_id,  network,  chain_id, kind)}/>
            }
            </Row>
          </div>
        }
      </Container>
    )
  }
}

const mapStateToProps = (state) => {
  var isFetching = true
  var showNewTaskForm = false
  var list = []


  if (state.tasks !== undefined && state.tasks !== null) {
    list = state.tasks.list
  }

  if (state.newtask !== undefined && state.newtask !== null) {
    showNewTaskForm = state.newtask.show
  }

  isFetching = false
  return {
    list,
    isFetching,
    showNewTaskForm
  }
}

export default connect(mapStateToProps)(Tasks)
