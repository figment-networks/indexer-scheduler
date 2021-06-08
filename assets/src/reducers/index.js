import { combineReducers } from 'redux'

import tasks from './tasks'
import lastdata from './lastdata'

const uiApp = combineReducers({
  tasks, lastdata
})

export default uiApp
