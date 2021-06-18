import {
  REQUEST_TASKS, RECEIVE_TASKS,INVALIDATE_TASKS,
  REQUEST_ENABLE_TASK, RECEIVE_ENABLE_TASK,
  REQUEST_DISABLE_TASK, RECEIVE_DISABLE_TASK,
  REQUEST_ADD_TASK, RECEIVE_ADD_TASK,
  UI_SHOW_ADDITIONAL_FIELDS,
} from '../actions'


const tasks = (state = {
    isFetching: false,
    isAdding: false,
    addingError: null,
    didInvalidate: false,
    enabling: false,
    disabling: false,
    operation_status: "",

    addTaskTypePicked: "",

    list: []
  }, action) => {
    switch (action.type) {
      case INVALIDATE_TASKS:
        return {
          ...state,
          didInvalidate: true
        }
      case REQUEST_ENABLE_TASK:
        return {
          ...state,
          enabling: true,
        }
      case RECEIVE_ENABLE_TASK:
        return {
          ...state,
          enabling: false,
          operation_status: action.status
        }
      case REQUEST_DISABLE_TASK:
        return {
          ...state,
          disabling: true,
        }
      case RECEIVE_DISABLE_TASK:
        return {
          ...state,
          disabling: false,
          operation_status: action.status
        }
      case REQUEST_ADD_TASK:
        return {
          ...state,
          isAdding: true,
        }
      case RECEIVE_ADD_TASK:
        return {
          ...state,
          isAdding: false,
          addingError: action.error
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
      case UI_SHOW_ADDITIONAL_FIELDS:
        return {
          ...state,
          addTaskTypePicked: action.kind
        }
      default:
        return state
    }
  }

  export default tasks


