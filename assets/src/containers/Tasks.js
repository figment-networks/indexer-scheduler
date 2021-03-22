import React, { Component } from 'react'
import PropTypes from 'prop-types'
import { connect } from 'react-redux'
import TaskList from '../components/TaskList'

import Button from 'react-bootstrap/Button'
import Container from 'react-bootstrap/Container'
import Row from 'react-bootstrap/Row'
import Col from 'react-bootstrap/Col'


import { fetchTasksIfNeeded, invalidateTasks, enableTask, disableTask  } from '../actions'
import { fetchLastData, invalidateLastdata  } from '../actions/lastdata'


class Tasks extends Component {
    static propTypes = {
      list: PropTypes.array.isRequired,
      isFetching: PropTypes.bool.isRequired,
      dispatch: PropTypes.func.isRequired
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
    const { list, isFetching } = this.props
    const isEmpty = list.length === 0
    return (
      <Container>
        <Row >
          <Col><h2>Task list</h2></Col>
          <Col xs={1}><Button variant="outline-dark" onClick={this.handleRefreshClick}>Refresh</Button></Col>
        </Row>
        <Row>
        {isEmpty
          ? (isFetching ? <h2>Loading...</h2> : "")
          : < TaskList
                  tasks={list}
                  loadTaskInformation={(task_id,  network,  chain_id, kind) => this.loadTaskInformation(task_id,  network,  chain_id, kind)}
                  enableTask={(task_id,  network,  chain_id, kind) => this.enableTask(task_id,  network,  chain_id, kind)}
                  disableTask={(task_id,  network,  chain_id, kind) => this.disableTask(task_id,  network,  chain_id, kind)}/>
        }
        </Row>
      </Container>
    )
  }
}

const mapStateToProps = (state) => {
  var isFetching = true
  var list = []

  if (state.tasks !== undefined && state.tasks !== null) {
    list = state.tasks.list
  }

  isFetching = false
  return {
    list,
    isFetching,
  }
}

export default connect(mapStateToProps)(Tasks)
