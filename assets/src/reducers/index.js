import { combineReducers } from 'redux'

import newtask from './newtask'
import tasks from './tasks'
import lastdata from './lastdata'

const uiApp = combineReducers({
  tasks, lastdata, newtask
})

export default uiApp
