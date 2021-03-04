import {
  REQUEST_TASKS, RECEIVE_TASKS,INVALIDATE_TASKS
} from '../actions'


const tasks = (state = {
    isFetching: false,
    didInvalidate: false,
    list: []
  }, action) => {
    switch (action.type) {
      case INVALIDATE_TASKS:
        return {
          ...state,
          didInvalidate: true
        }
      case REQUEST_TASKS:
        return {
          ...state,
          isFetching: true,
          didInvalidate: false
        }
      case RECEIVE_TASKS:
        return {
          ...state,
          isFetching: false,
          didInvalidate: false,
          list: action.tasks,
          lastUpdated: action.receivedAt
        }
      default:
        return state
    }
  }

  export default tasks
