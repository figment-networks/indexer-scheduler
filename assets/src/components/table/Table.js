import React from 'react'
import PropTypes from 'prop-types'

import BootstrapTable from 'react-bootstrap-table-next'
import 'react-bootstrap-table-next/dist/react-bootstrap-table2.min.css'

const rowClasses = (row, rowIndex) => {
  if (rowIndex % 2 === 0) {
    return 'row-even'
  }

  return 'row-odd'
}

class Table extends React.Component {
    static propTypes = {
      columns: PropTypes.array.isRequired,
      data: PropTypes.array.isRequired,
      tableName: PropTypes.string.isRequired
    }

    render () {
      const { columns, data, tableName } = this.props
      return (
            <BootstrapTable bootstrap4 keyField={tableName} className="table table-bordered table-responsive" columns={ columns } data={ data } bodyClasses="row-body" rowClasses={rowClasses}/>
      )
    }
}

export default Table
