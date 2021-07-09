import {
  UI_HIDE_LOGS, UI_SHOW_LOGS
} from '../actions'

const showlogs = (state = {
  show: false
}, action) => {
  switch (action.type) {
    case UI_HIDE_LOGS:
      return {
        show: false
      }
    case UI_SHOW_LOGS:
      return {
        show: true
      }
    default:
      return state
  }
}

export default showlogs
