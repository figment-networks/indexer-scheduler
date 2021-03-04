import React, { Component } from 'react'
import PropTypes from 'prop-types'
import { connect } from 'react-redux'

import Button from 'react-bootstrap/Button'
import Container from 'react-bootstrap/Container'
import Row from 'react-bootstrap/Row'


import LastdataList from '../components/LastdataList'
import { fetchLastData, invalidateLastdata  } from '../actions/lastdata'


class LastData extends Component {
    static propTypes = {
      list: PropTypes.array.isRequired,
      isFetching: PropTypes.bool.isRequired,

      task_id: PropTypes.string.isRequired,
      network: PropTypes.string.isRequired,
      chain_id: PropTypes.string.isRequired,
      kind: PropTypes.string.isRequired,

      dispatch: PropTypes.func.isRequired,
    }

/*
  componentDidMount() {
    //this.props.dispatch(fetchLastdataIfNeeded())
  }

  componentDidUpdate(prevProps) {
    if (prevProps.list !== this.props.list) {
    }
  }
*/


  handleRefreshClick() {
    const { dispatch } = this.props
    dispatch(invalidateLastdata())
    dispatch(fetchLastData(this.props.task_id, this.props.network, this.props.chain_id, this.props.kind, 100, 0))

  }


  render() {
    const { list, isFetching } = this.props
    const isEmpty = list.length === 0
    return (
      <Container>
      <Row>
        <Button variant="outline-dark" onClick={(e) => this.handleRefreshClick()}>Refresh</Button>
      </Row>
      <Row>
      {isEmpty
        ? (isFetching ? <h2>Loading...</h2> : <h2>Empty.</h2>)
        : < LastdataList lastdatas={list} />
      }
      </Row>
    </Container>
    )
  }
}

const mapStateToProps = (state) => {
  const isFetching = false;
  var list = [];
  let task_id = state.lastdata.task_id;
  let network = state.lastdata.network;
  let chain_id = state.lastdata.chain_id;
  let kind = state.lastdata.kind;

  if (state.lastdata !== undefined && state.lastdata !== null) {
    list = state.lastdata.list
  }


  return {
    list,
    isFetching,
    task_id,
    network,
    chain_id,
    kind,
  }
}

export default connect(mapStateToProps)(LastData)
