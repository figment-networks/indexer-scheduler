export const REQUEST_TASKS = 'REQUEST_TASKS'
export const RECEIVE_TASKS = 'RECEIVE_TASKS'
export const INVALIDATE_TASKS = 'INVALIDATE_TASKS'

export const RECEIVE_GET_LAST_HEIGHTS = 'RECEIVE_GET_LAST_HEIGHTS'

export const DISABLE_AUTO_REFRESH = 'DISABLE_AUTO_REFRESH'
export const ENABLE_AUTO_REFRESH = 'ENABLE_AUTO_REFRESH'

export const REQUEST_GET_ALL_LATEST = 'REQUEST_GET_ALL_LATEST'
export const RECEIVE_GET_ALL_LATEST = 'RECEIVE_GET_ALL_LATEST'

export const REQUEST_DELETE_TASK = 'REQUEST_DELETE_TASK'
export const RECEIVE_DELETE_TASK = 'RECEIVE_DELETE_TASK'

export const REQUEST_DISABLE_TASK = 'REQUEST_DISABLE_TASK'
export const RECEIVE_DISABLE_TASK = 'RECEIVE_DISABLE_TASK'

export const REQUEST_ENABLE_TASK = 'REQUEST_ENABLE_TASK'
export const RECEIVE_ENABLE_TASK = 'RECEIVE_ENABLE_TASK'

export const REQUEST_ADD_TASK = 'REQUEST_ADD_TASK'
export const RECEIVE_ADD_TASK = 'RECEIVE_ADD_TASK'

export const UI_SHOW_ADDITIONAL_FIELDS = 'UI_SHOW_ADDITIONAL_FIELDS'

export const UI_HIDE_NEW_TASK = 'UI_HIDE_NEW_TASK'
export const UI_SHOW_NEW_TASK = 'UI_SHOW_NEW_TASK'

let initial = true

export const invalidateTasks = () => ({
  type: INVALIDATE_TASKS
})

export const requestTasks = () => ({
  type: REQUEST_TASKS
})

export const receiveTasks = (json) => ({
  type: RECEIVE_TASKS,
  tasks: json.map(task => {
    task.config = task.config ? JSON.stringify(task.config, null, 4) : ''
    return task
  }),
  receivedAt: Date.now()
})

const fetchTasks = () => dispatch => {
  dispatch(requestTasks())
  return fetch('http://0.0.0.0:8889/scheduler/core/list')
    .then(response => response.json())
    .then(json => dispatch(receiveTasks(json)))
}

const shouldFetchTasks = (state) => {
  const tasks = state.tasks
  if (!tasks || tasks === undefined) {
    return true
  }

  if (tasks.isFetching) {
    return false
  }

  return tasks.didInvalidate
}

export const fetchTasksIfNeeded = () => (dispatch, getState) => {
  if (initial || shouldFetchTasks(getState())) {
    initial = false
    return dispatch(fetchTasks())
  }
}

export const requestGetAllLatest = () => ({
  type: REQUEST_GET_ALL_LATEST
})

export const receiveGetAllLatest = (json) => ({
  type: RECEIVE_GET_ALL_LATEST,
  status: json.error ? json.error : '',
  receivedAt: Date.now()
})

export const requestDeleteTask = () => ({
  type: REQUEST_DELETE_TASK
})

export const receiveDeleteTask = (json) => dispatch => {
  dispatch(invalidateTasks())
  dispatch(fetchTasksIfNeeded())
  return {
    type: RECEIVE_DELETE_TASK,
    status: json.error ? json.error : '',
    receivedAt: Date.now()
  }
}

export const requestDisableTask = () => ({
  type: REQUEST_DISABLE_TASK
})

export const receiveDisableTask = (json) => dispatch => {
  dispatch(invalidateTasks())
  dispatch(fetchTasksIfNeeded())
  return {
    type: RECEIVE_DISABLE_TASK,
    status: json.error ? json.error : '',
    receivedAt: Date.now()
  }
}

export const requestEnableTask = () => ({
  type: REQUEST_ENABLE_TASK
})

export const receiveEnableTask = (json) => dispatch => {
  dispatch(invalidateTasks())
  dispatch(fetchTasksIfNeeded())
  return {
    type: RECEIVE_ENABLE_TASK,
    status: json.error ? json.error : '',
    receivedAt: Date.now()
  }
}

export const disableAutoRefresh = () => ({
  type: DISABLE_AUTO_REFRESH
})

export const enableAutoRefresh = () => ({
  type: ENABLE_AUTO_REFRESH
})

export const enableTask = (taskID) => dispatch => {
  dispatch(requestEnableTask(taskID))
  return fetch('http://0.0.0.0:8889/scheduler/core/enable/' + taskID, { method: 'GET' })
    .then(response => response.json())
    .then(json => dispatch(receiveEnableTask(json)))
}

export const disableTask = (taskID) => dispatch => {
  dispatch(requestDisableTask(taskID))
  return fetch('http://0.0.0.0:8889/scheduler/core/disable/' + taskID, { method: 'GET' })
    .then(response => response.json())
    .then(json => dispatch(receiveDisableTask(json)))
}

export const deleteTask = (taskID) => dispatch => {
  dispatch(requestDeleteTask(taskID))
  return fetch('http://0.0.0.0:8889/scheduler/core/deleteTask/' + taskID, { method: 'GET' })
    .then(response => response.json())
    .then(json => dispatch(receiveDeleteTask(json)))
}

export const receiveGetLastHeights = (json) => ({
  type: RECEIVE_GET_LAST_HEIGHTS,
  list: json
})

export const getLastHeights = () => dispatch => {
  return fetch('http://0.0.0.0:8889/scheduler/core/getLastHeights', { method: 'GET' })
    .then(response => response.json())
    .then(json => dispatch(receiveGetLastHeights(json)))
}

export const receiveAddTask = (json) => dispatch => {
  dispatch(invalidateTasks())
  dispatch(fetchTasksIfNeeded())
  return {
    type: RECEIVE_ADD_TASK,
    error: json.error ? json.error : '',
    receivedAt: Date.now()
  }
}

export const requestAddTask = () => ({
  type: REQUEST_ADD_TASK
})

export const addTask = (addTaskParam) => dispatch => {
  dispatch(requestAddTask())
  return fetch('http://0.0.0.0:8889/scheduler/core/addTask/', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(addTaskParam)
  })
    .then(response => response.json())
    .then(json => dispatch(receiveAddTask(json)))
}

export const changeAdditionalFields = (kind) => ({
  type: UI_SHOW_ADDITIONAL_FIELDS,
  kind
})

export const hideNewTask = () => ({
  type: UI_HIDE_NEW_TASK
})

export const showNewTask = () => ({
  type: UI_SHOW_NEW_TASK
})
