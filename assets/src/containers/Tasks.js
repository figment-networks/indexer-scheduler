import React, { Component } from 'react'
import PropTypes from 'prop-types'
import { connect } from 'react-redux'
import TaskList from '../components/TaskList'

import Button from 'react-bootstrap/Button'
import Container from 'react-bootstrap/Container'
import Row from 'react-bootstrap/Row'


import { fetchTasksIfNeeded, invalidateTasks  } from '../actions'
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

  render() {
    const { list, isFetching } = this.props
    const isEmpty = list.length === 0
    return (
      <Container>
        <Row>
          <Button variant="outline-dark" onClick={this.handleRefreshClick}>Refresh</Button>
        </Row>
        <Row>
        {isEmpty
          ? (isFetching ? <h2>Loading...</h2> : <h2>Empty.</h2>)
          : < TaskList tasks={list} loadTaskInformation={(task_id,  network,  chain_id, kind) => this.loadTaskInformation(task_id,  network,  chain_id, kind)} />
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
