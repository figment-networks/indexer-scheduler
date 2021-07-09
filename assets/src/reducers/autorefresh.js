import {
  ENABLE_AUTO_REFRESH, DISABLE_AUTO_REFRESH
} from '../actions'

const autorefresh = (state = {
  show: false
}, action) => {
  switch (action.type) {
    case ENABLE_AUTO_REFRESH:
      return {
        ...state,
        show: true
      }
    case DISABLE_AUTO_REFRESH:
      return {
        ...state,
        show: false
      }
    default:
      return state
  }
}

export default autorefresh
