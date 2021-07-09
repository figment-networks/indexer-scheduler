import React, { Component } from 'react'
import PropTypes from 'prop-types'
import { connect } from 'react-redux'
import NewTask from './NewTask'
import LastData from '../log/LastData'
import Table from '../table/Table'

import Button from 'react-bootstrap/Button'

import { fetchTasksIfNeeded, invalidateTasks, enableTask, deleteTask, disableTask, showNewTask, hideNewTask } from '../../actions'
import { fetchLastData, invalidateLastdata } from '../../actions/lastdata'
import Header from '../table/Header'

import { dangerStyle, primaryStyle, successStyle } from '../../style/button'
import LastHeights from './LastHeights'

const taskColumns = [{
  headerStyle: { width: '100px' },
  dataField: 'task_id',
  sort: true,
  text: 'Task ID'
}, {
  headerStyle: { width: '110px' },
  dataField: 'network',
  sort: true,
  text: 'Network'
}, {
  headerStyle: { width: '110px' },
  dataField: 'chain_id',
  sort: true,
  text: 'Chain ID'
}, {
  headerStyle: { width: '100px' },
  dataField: 'kind',
  sort: true,
  text: 'Kind'
}, {
  headerStyle: { width: '130px' },
  dataField: 'duration',
  sort: true,
  text: 'Duration'
}, {
  headerStyle: { width: '100px' },
  dataField: 'status',
  sort: true,
  text: 'Status'
}, {
  headerStyle: { width: '100px' },
  dataField: 'enabled',
  sort: true,
  text: 'Enabled'
}, {
  headerStyle: { width: '230px' },
  dataField: 'config',
  sort: true,
  text: 'Config'
}, {
  headerStyle: { width: '160px' },
  dataField: 'buttons',
  style: { wordWrap: 'normal', margin: '3px', padding: '3px' },
  text: ''
}]

class Tasks extends Component {
  static propTypes = {
    list: PropTypes.array.isRequired,
    isFetching: PropTypes.bool.isRequired,
    dispatch: PropTypes.func.isRequired,
    showNewTaskForm: PropTypes.bool.isRequired
  }

  componentDidMount () {
    const { dispatch } = this.props
    dispatch(fetchTasksIfNeeded())
  }

  componentDidUpdate (prevProps) {
    const { dispatch, list } = this.props
    if (prevProps.list !== list) {
      dispatch(fetchTasksIfNeeded())
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

  loadTaskInformation (taskID, network, chainID, kind) {
    const { dispatch } = this.props
    dispatch(invalidateLastdata(taskID))
    dispatch(fetchLastData(taskID, network, chainID, kind, 100, 0))
  }

  clickEnableTask (taskID, network, chainID, kind) {
    const { dispatch } = this.props
    dispatch(enableTask(taskID, network, chainID, kind))
  }

  clickDisableTask (taskID, network, chainID, kind) {
    const { dispatch } = this.props
    dispatch(disableTask(taskID, network, chainID, kind))
  }

  clickDeleteTask (taskID) {
    const { dispatch } = this.props
    dispatch(deleteTask(taskID))
  }

  formatRows (tasks) {
    return tasks.map(task => {
      task.buttons = <div className="tableButtons">
            <Button style={primaryStyle} onClick={(e) => this.loadTaskInformation(task.task_id, task.network, task.chain_id, task.kind)}><span>LOGS</span></Button>
            {task.enabled
              ? <Button style={dangerStyle} onClick={(e) => this.clickDisableTask(task.id, task.network, task.chain_id, task.kind)}>STOP</Button>
              : <Button style={successStyle} onClick={(e) => this.clickEnableTask(task.id, task.network, task.chain_id, task.kind)}>START</Button>}
            <Button style={dangerStyle} onClick={(e) => this.clickDeleteTask(task.task_id)}>DELETE</Button>
        </div>
      return task
    })
  }

  render () {
    const { list, isFetching, showNewTaskForm } = this.props
    const isEmpty = (list === null || list === undefined || list.length === 0)
    return (
      <div>
        <LastHeights/>
        {showNewTaskForm
          ? <div className="center">
          <div className="box">
              <h2 className="top-title">Create a new task</h2>
              <Button className="box-button" style={dangerStyle} onClick={this.handleHideTaskClick}>X</Button>
              <NewTask/>
          </div>
        </div>
          : null }
          <div>
            <Header title="Task list"
            handleNewTaskClick={this.handleNewTaskClick}
            handleRefreshClick={this.handleRefreshClick}
            />
            {isEmpty
              ? (isFetching ? <h2>Loading...</h2> : '')
              : <Table columns={taskColumns} data={this.formatRows(list)} tableName="task_id"/>
            }
          </div>
            <LastData/>
      </div>
    )
  }
}

const mapStateToProps = (state) => {
  let isFetching = true
  let showNewTaskForm = false
  let list = []

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
