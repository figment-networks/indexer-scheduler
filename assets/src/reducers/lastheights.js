import {
  RECEIVE_GET_LAST_HEIGHTS
} from '../actions'

const lastheights = (state = {
  list: []
}, action) => {
  switch (action.type) {
    case RECEIVE_GET_LAST_HEIGHTS:
      console.log('actionsss', action)
      return {
        ...state,
        list: action.list
      }
    default:
      return state
  }
}

export default lastheights
