import React, { Component } from 'react'
import PropTypes from 'prop-types'
import { connect } from 'react-redux'

import Button from 'react-bootstrap/Button'
import Container from 'react-bootstrap/Container'
import Row from 'react-bootstrap/Row'
import Form from 'react-bootstrap/Form'

import { addTask, changeAdditionalFields, hideNewTask } from '../../actions/index'

class NewTask extends Component {
  static propTypes = {
    isAdding: PropTypes.bool.isRequired,
    dispatch: PropTypes.func.isRequired,
    addTaskTypePicked: PropTypes.string.isRequired
  }

  taskKindChange (selectedOption) {
    const { dispatch } = this.props
    dispatch(changeAdditionalFields(this.kindVal.value))
  }

  handleSubmit (e) {
    e.preventDefault()
    const { dispatch } = this.props

    const addTaskPayload = {
      task_id: this.taskIDVal.value,
      kind: this.kindVal.value,
      network: this.networkVal.value,
      chain_id: this.chainIDVal.value,
      interval: this.intervalVal.value
    }

    if (this.kindVal.value === 'syncrange') {
      addTaskPayload.config = {
        height_from: this.heightFromVal.value,
        height_to: this.heightToVal.value
      }
    }

    dispatch(addTask(addTaskPayload))
    dispatch(hideNewTask())
  }

  render () {
    return (
      <div>
        <Row style={{ padding: 20 }}>
        <Form onSubmit={(e) => this.handleSubmit(e)}>
          <Form.Group controlId="newTaskTaskID">
            <Form.Label>TaskID</Form.Label>
            <Form.Control type="text" placeholder="Enter task_id" ref={node => (this.taskIDVal = node)} />
            <Form.Text className="text-muted">
              Unique taskID. This value has to be unique if you wanna run few tasks simultaneously.
            </Form.Text>
          </Form.Group>

          <Form.Group controlId="newTaskNetwork">
            <Form.Label>Network</Form.Label>
            <Form.Control type="text" placeholder="Enter network name" ref={node => (this.networkVal = node)} />
            <Form.Text className="text-muted">
              Name of the network used in project. eg. `skale`, `cosmos`,`terra`
            </Form.Text>
          </Form.Group>

          <Form.Group controlId="newTaskChainID">
            <Form.Label>ChainID</Form.Label>
            <Form.Control type="text" placeholder="Enter chain_id" ref={node => (this.chainIDVal = node)} />
            <Form.Text className="text-muted">
              Name of the chain_id used in project. eg. `mainnet`, `cosmoshub-4`
            </Form.Text>
          </Form.Group>

          <Form.Group controlId="newTaskInterval">
            <Form.Label>Interval</Form.Label>
            <Form.Control type="text" placeholder="Enter interval" ref={node => (this.intervalVal = node)} />
            <Form.Text className="text-muted">
              Interval of the least time between runs. Usual go parsing is applied, so 10s, 2m, 4w
            </Form.Text>
          </Form.Group>

          <Form.Group controlId="newTaskKind">
            <Form.Label>Type (runner)</Form.Label>
            <Form.Control as="select" onChange={(e) => this.taskKindChange(e)} ref={node => (this.kindVal = node)} >
              <option>lastdata</option>
              <option>syncrange</option>
            </Form.Control>
          </Form.Group>

          {this.props.addTaskTypePicked === 'syncrange'
            ? <Container style={{ padding: 0 }}>
              <h3>Sync Range params:</h3>
              <hr />
              <Form.Group controlId="newTaskHeightFrom">
                  <Form.Label>Height From</Form.Label>
                  <Form.Control type="text" placeholder=" Enter starting height" ref={node => (this.heightFromVal = node)} />
                  <Form.Text className="text-muted">
                    Height that task will start with
                  </Form.Text>
                </Form.Group>
                <Form.Group controlId="newTaskHeightTo">
                  <Form.Label>Height To</Form.Label>
                  <Form.Control type="text" placeholder="Enter finish height" ref={node => (this.heightToVal = node)} />
                  <Form.Text className="text-muted">
                    Height that task will finish on
                  </Form.Text>
                </Form.Group>
              </Container>
            : ''}

          <Button variant="outline-primary" size="lg" type="submit" block>
            Submit
          </Button>

        </Form>
        </Row>
      </div>
    )
  }
}

const mapStateToProps = (state) => {
  const isAdding = false
  const addTaskTypePicked = state.tasks.addTaskTypePicked

  return {
    isAdding,
    addTaskTypePicked
  }
}

export default connect(mapStateToProps)(NewTask)
