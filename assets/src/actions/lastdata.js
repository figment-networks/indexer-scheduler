export const REQUEST_LASTDATA = 'REQUEST_LASTDATA'
export const RECEIVE_LASTDATA = 'RECEIVE_LASTDATA'
export const INVALIDATE_LASTDATA = 'INVALIDATE_LASTDATA'

export const invalidateLastdata = (taskID) => ({
  type: INVALIDATE_LASTDATA,
  task_id: taskID
})

export const requestLastdata = (taskID, network, chainID, kind) => ({
  type: REQUEST_LASTDATA,
  task_id: taskID,
  chain_id: chainID,
  network: network,
  kind: kind
})

export const receiveLastdata = (json) => ({
  type: RECEIVE_LASTDATA,
  lastdata: json,
  receivedAt: Date.now()
})

export const fetchLastData = (taskID, network, chainID, kind, limit, offset) => dispatch => {
  dispatch(requestLastdata(taskID, network, chainID, kind))
  return fetch('http://0.0.0.0:8889/scheduler/runner/' + kind + '/listRunning', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    mode: 'cors',
    body: JSON.stringify({
      kind,
      network,
      taskID,
      chainID,
      limit,
      offset
    })
  })
    .then(response => response.json())
    .then(json => dispatch(receiveLastdata(json, taskID, network, chainID, kind)))
}
