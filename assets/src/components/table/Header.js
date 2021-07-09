import React, { Component } from 'react'
import PropTypes from 'prop-types'
import { connect } from 'react-redux'

import Button from 'react-bootstrap/Button'

import { disableAutoRefresh, enableAutoRefresh } from '../../actions'

import { buttonStyle, primaryStyle, successStyle } from '../../style/button'

const defaultButton = {
  ...buttonStyle,
  fontSize: '1rem'
}

const autorefreshOFF = {
  ...defaultButton,
  backgroundColor: '#a59797',
  width: '170px'
}

const autorefreshON = {
  ...successStyle,
  ...defaultButton,
  width: '170px'
}

const primary = {
  ...primaryStyle,
  ...defaultButton,
  width: '80px'
}

const success = {
  ...successStyle,
  ...defaultButton,
  width: '180px'
}

class Header extends Component {
    static propTypes = {
      desc: PropTypes.string,
      title: PropTypes.string.isRequired,
      dispatch: PropTypes.func.isRequired,

      refreshInterval: PropTypes.bool,
      autoRefresh: PropTypes.bool,

      handleNewTaskClick: PropTypes.func,
      handleRefreshClick: PropTypes.func.isRequired,
      refreshLatestData: PropTypes.func
    }

    componentDidMount () {
      setInterval(this.refreshIfNeeded, 60 * 1000)
    }

    refreshIfNeeded = () => {
      const { autoRefresh, refreshLatestData, refreshInterval } = this.props
      if (autoRefresh && refreshInterval) {
        refreshLatestData()
      }
    }

    clickEnableAutoRefresh = e => {
      e.preventDefault()

      const { dispatch } = this.props
      dispatch(enableAutoRefresh())
    }

    clickDisableAutoRefresh = e => {
      e.preventDefault()

      const { dispatch } = this.props
      dispatch(disableAutoRefresh())
    }

    render () {
      const { autoRefresh, desc, title, handleNewTaskClick, handleRefreshClick, refreshInterval } = this.props

      return <div className="header">
            <h2>{title}</h2>
            <span >{desc}</span>

            <div className="headerButtons">
            {
                handleNewTaskClick
                  ? <Button style={success} onClick={handleNewTaskClick}>CREATE A NEW TASK</Button>
                  : null
            }
            {
                refreshInterval
                  ? autoRefresh
                    ? <Button style={autorefreshON} onClick={this.clickDisableAutoRefresh}>AUTO-REFRESH ON</Button>
                    : <Button style={autorefreshOFF} onClick={this.clickEnableAutoRefresh}>AUTO-REFRESH OFF</Button>
                  : null
            }
            <Button style={primary} onClick={handleRefreshClick}>REFRESH</Button>
            </div>

        </div>
    }
}

const mapStateToProps = (state) => {
  let autoRefresh = false

  if (state.autorefresh !== undefined && state.autorefresh !== null) {
    autoRefresh = state.autorefresh.show
  }

  return {
    autoRefresh
  }
}

export default connect(mapStateToProps)(Header)
