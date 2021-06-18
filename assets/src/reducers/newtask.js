import {
    UI_HIDE_NEW_TASK, UI_SHOW_NEW_TASK
} from '../actions'
  
  
const newtask = (state = {
    show: false,
}, action) => {
  switch (action.type) {
    case UI_HIDE_NEW_TASK:
      return {
        show: false
      }
    case UI_SHOW_NEW_TASK:
      return {
        show: true
      }
    default:
      return state
  }
}

export default newtask
  