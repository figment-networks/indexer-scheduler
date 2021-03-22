export const REQUEST_TASKS = 'REQUEST_TASKS'
export const RECEIVE_TASKS = 'RECEIVE_TASKS'
export const INVALIDATE_TASKS = 'INVALIDATE_TASKS'

export const REQUEST_DISABLE_TASK = 'REQUEST_DISABLE_TASKS'
export const RECEIVE_DISABLE_TASK = 'RECEIVE_DISABLE_TASKS'

export const REQUEST_ENABLE_TASK = 'REQUEST_ENABLE_TASKS'
export const RECEIVE_ENABLE_TASK = 'RECEIVE_ENABLE_TASKS'


export const REQUEST_ADD_TASK = 'REQUEST_ADD_TASKS'
export const RECEIVE_ADD_TASK = 'RECEIVE_ADD_TASKS'

export const UI_SHOW_ADDITIONAL_FIELDS = 'UI_SHOW_ADDITIONAL_FIELDS'



var initial = true

export const invalidateTasks = () => ({
  type: INVALIDATE_TASKS
})

export const requestTasks = () => ({
  type: REQUEST_TASKS
})

export const receiveTasks = (json) => ({
  type: RECEIVE_TASKS,
  tasks: json,
  receivedAt: Date.now()
})


const fetchTasks = () => dispatch => {
  dispatch(requestTasks())
  return fetch(`/scheduler/core/list`)
    .then(response => response.json())
    .then(json => dispatch(receiveTasks(json)))
}

const shouldFetchTasks = (state) => {
  const tasks = state.tasks;
  if (!tasks || tasks === undefined  ) {
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

export const requestDisableTask = () => ({
  type: REQUEST_DISABLE_TASK
})

export const receiveDisableTask = (json) => ({
  type: RECEIVE_DISABLE_TASK,
  status: json.error ? json.error : "" ,
  receivedAt: Date.now()
})

export const requestEnableTask = () => ({
  type: REQUEST_ENABLE_TASK
})

export const receiveEnableTask = (json) => ({
  type: RECEIVE_ENABLE_TASK,
  status: json.error ? json.error : "" ,
  receivedAt: Date.now()
})

export const enableTask = (task_id) => dispatch => {
  dispatch(requestEnableTask(task_id))
  return fetch(`/scheduler/core/enable/`+ task_id, {method: 'GET'})
    .then(response => response.json())
    .then(json => dispatch(receiveEnableTask(json)))
}

export const disableTask = (task_id) => dispatch => {
  dispatch(requestDisableTask(task_id))
  return fetch(`/scheduler/core/disable/`+ task_id, {method: 'GET'})
    .then(response => response.json())
    .then(json => dispatch(receiveDisableTask(json)))
}

////// NEW TASK

export const requestAddTask = () => ({
  type: REQUEST_ADD_TASK
})

export const receiveAddTask = (json) => ({
  type: RECEIVE_ADD_TASK,
  error: json.error ? json.error : "" ,
  receivedAt: Date.now()
})

export const addTask = (addTaskParam) => dispatch => {
  dispatch(requestAddTask())
  return fetch(`/scheduler/core/addTask/`, {
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
