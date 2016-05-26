import React, {Component, PropTypes} from 'react';
import moment from 'moment';
class DataTable extends Component {
    static propTypes = {
        data: PropTypes.arrayOf(PropTypes.object).isRequired
    }

    render() {

        return (
            <table className="table table-striped">
                <thead>
                    <tr>
                        <th>Timestamp</th>
                        <th>Data</th>
                    </tr>
                </thead>
                <tbody>
                    {this.props.data.map((s) => {
                        return (
                            <tr key={s.t}>
                                <td>{moment(new Date(s.t * 1000)).calendar()}</td>
                                <td>{JSON.stringify(s.d)}</td>
                            </tr>
                        );
                    })}
                </tbody>
            </table>
        );
    }
}
export default DataTable;
