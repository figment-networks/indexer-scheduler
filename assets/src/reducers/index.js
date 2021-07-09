import { combineReducers } from 'redux'

import autorefresh from './autorefresh'
import lastheights from './lastheights'
import newtask from './newtask'
import tasks from './tasks'
import lastdata from './lastdata'

const uiApp = combineReducers({
  tasks, lastdata, newtask, autorefresh, lastheights
})

export default uiApp
