export const REQUEST_LASTDATA = 'REQUEST_LASTDATA'
export const RECEIVE_LASTDATA = 'RECEIVE_LASTDATA'
export const INVALIDATE_LASTDATA = 'INVALIDATE_LASTDATA'


export const invalidateLastdata = () => ({
  type: INVALIDATE_LASTDATA
})


export const requestLastdata = (task_id, network, chain_id, kind,) => ({
  type: REQUEST_LASTDATA,
  task_id: task_id,
  chain_id: chain_id,
  network: network,
  kind: kind,
})

export const receiveLastdata = (json) => ({
  type: RECEIVE_LASTDATA,
  lastdata: json,
  receivedAt: Date.now()
})

export const fetchLastData = ( task_id, network, chain_id, kind, limit, offset) => dispatch => {
  dispatch(requestLastdata(task_id, network, chain_id, kind))
  return fetch(`/scheduler/runner/`+ kind +`/listRunning`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
	    kind,
	    network,
	    task_id,
      chain_id,
      limit,
      offset
    })
  })
    .then(response => response.json())
    .then(json => dispatch(receiveLastdata(json)))
}


