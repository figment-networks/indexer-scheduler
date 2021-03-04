export const REQUEST_TASKS = 'REQUEST_TASKS'
export const RECEIVE_TASKS = 'RECEIVE_TASKS'
export const INVALIDATE_TASKS = 'INVALIDATE_TASKS'

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

