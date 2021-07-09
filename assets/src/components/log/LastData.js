import React, { Component } from 'react'
import PropTypes from 'prop-types'
import { connect } from 'react-redux'

import LogList from './LogList'
import { fetchLastData, invalidateLastdata } from '../../actions/lastdata'
import Header from '../table/Header'

class LastData extends Component {
    static propTypes = {
      list: PropTypes.array.isRequired,
      isFetching: PropTypes.bool.isRequired,

      taskID: PropTypes.string.isRequired,
      network: PropTypes.string.isRequired,
      chainID: PropTypes.string.isRequired,
      kind: PropTypes.string.isRequired,

      dispatch: PropTypes.func.isRequired
    }

  handleRefreshClick = e => {
    e.preventDefault()
    const { dispatch } = this.props
    dispatch(invalidateLastdata(this.props.taskID))
    dispatch(fetchLastData(this.props.taskID, this.props.network, this.props.chainID, this.props.kind, 100, 0))
  }

  refreshLatestData = () => {
    const { dispatch, taskID, network, chainID, kind } = this.props
    dispatch(invalidateLastdata(taskID))
    dispatch(fetchLastData(taskID, network, chainID, kind, 100, 0))
  }

  render () {
    const { list, isFetching, network, chainID, kind, taskID } = this.props
    const isEmpty = (list === null || list.length === 0)
    const desc = <div><b>task_id</b>: {taskID} <b>kind</b>: {kind} <b>chain_id</b>: {chainID} <b>network</b>: {network}</div>
    return (
      <div>
        {isEmpty
          ? ''
          : <Header desc={desc}
            title="Last Data"
            handleRefreshClick={this.handleRefreshClick}
            refreshLatestData={this.refreshLatestData}
            refreshInterval={true}/>
        }
      {isEmpty
        ? (isFetching ? <h2>Loading...</h2> : '')
        : <LogList logs={list}/>
      }
    </div>
    )
  }
}

const mapStateToProps = (state) => {
  const isFetching = false
  let list = []
  const taskID = state.lastdata.task_id
  const network = state.lastdata.network
  const chainID = state.lastdata.chain_id
  const kind = state.lastdata.kind

  if (state.lastdata !== undefined && state.lastdata !== null) {
    list = state.lastdata.list
  }

  return {
    list,
    isFetching,
    taskID,
    network,
    chainID,
    kind
  }
}

export default connect(mapStateToProps)(LastData)
