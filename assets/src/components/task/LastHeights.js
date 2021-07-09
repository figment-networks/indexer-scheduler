import React from 'react'
import PropTypes from 'prop-types'
import { connect } from 'react-redux'

import { getLastHeights } from '../../actions'
import Table from '../table/Table'

const lastHeightsColumns = [{
  headerStyle: { width: '110px' },
  dataField: 'network',
  text: 'Network'
}, {
  headerStyle: { width: '110px' },
  dataField: 'chain_id',
  text: 'Chain ID'
}, {
  headerStyle: { width: '100px' },
  dataField: 'height',
  text: 'Height'
}]

class LastHeights extends React.Component {
    static propTypes = {
      dispatch: PropTypes.func.isRequired,
      list: PropTypes.array.isRequired
    }

    componentDidMount () {
      const { dispatch } = this.props
      dispatch(getLastHeights())
    }

    render () {
      const { list } = this.props
      return (
        list ? <Table columns={lastHeightsColumns} data={list} tableName="id"/> : null
      )
    }
}

const mapStateToProps = (state) => {
  let list = []

  if (state.lastheights !== undefined && state.lastheights !== null) {
    list = state.lastheights.list
  }

  return {
    list
  }
}

export default connect(mapStateToProps)(LastHeights)
