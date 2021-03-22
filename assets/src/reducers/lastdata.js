import {
  RECEIVE_LASTDATA, REQUEST_LASTDATA, INVALIDATE_LASTDATA
} from '../actions/lastdata'


const lastdata = (state = {
    isFetching: false,
    didInvalidate: false,
    task_id: "",
    chain_id: "",
    network: "",
    kind: "",
    list: []
  }, action) => {
    switch (action.type) {
      case INVALIDATE_LASTDATA:
        return {
          ...state,
          didInvalidate: true
        }
      case REQUEST_LASTDATA:
        return {
          ...state,
          task_id: action.task_id,
          chain_id: action.chain_id,
          network: action.network,
          kind: action.kind,
          isFetching: true,
          didInvalidate: false
        }
      case RECEIVE_LASTDATA:
        for (const ld of action.lastdata) {
          if (ld.error !== undefined && ld.error !== "") {
            ld.error = atob(ld.error)
          }
        }
        return {
          ...state,
          isFetching: false,
          didInvalidate: false,
          list: action.lastdata,
          lastUpdated: action.receivedAt
        }
      default:
        return state
    }
  }

  export default lastdata
