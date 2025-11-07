import React, {Component} from 'react';
import {Button, Space, Table, Tooltip} from "antd";
import {arrayEquals, renderSize} from "../utils/utils.jsx";
import request from "../common/request";
import {FormattedMessage} from "react-intl";

class TopicPartition extends Component {

    state = {
        loading: false,
        items: [],
        clusterId: undefined,
        topic: undefined
    }

    componentDidMount() {
        let clusterId = this.props.clusterId;
        let topic = this.props.topic;
        this.setState({
            clusterId: clusterId,
            topic: topic
        })
        this.props.onRef(this);
        this.loadItems(clusterId, topic);
    }

    async loadItems(clusterId, topic) {
        this.setState({
            loading: true
        })
        let response = await request.get(`/topics/${topic}/partitions?clusterId=${clusterId}`);
        let items = response && Array.isArray(response.data) ? response.data : [];
        this.setState({
            items: items,
            loading: false
        })
    }

    refresh() {
        if (this.state.clusterId && this.state.topic) {
            this.loadItems(this.state.clusterId, this.state.topic)
        }
    }

    render() {

        const columns = [{
            title: 'Partition',
            dataIndex: 'partition',
            key: 'partition'
        }, {
            title: 'Leader',
            dataIndex: 'leader',
            key: 'leader',
            defaultSortOrder: 'ascend',
            render: (leader, record, index) => {
                if (!leader) {
                    return '-';
                }
                const host = leader['host'] || '';
                const port = leader['port'] || '';
                return <Tooltip key={'leader-' + host} title={`${host}:${port}`}>
                    <Button key={host} size='small'>{leader['id']}</Button>
                </Tooltip>
            }
        }, {
            title: 'Beginning Offset',
            dataIndex: 'beginningOffset',
            key: 'beginningOffset',
            sorter: (a, b) => a['beginningOffset'] - b['beginningOffset'],
        }, {
            title: 'End Offset',
            dataIndex: 'endOffset',
            key: 'endOffset',
            sorter: (a, b) => a['endOffset'] - b['endOffset'],
        }, {
                title: 'Log Size',
                dataIndex: 'y',
                key: 'y',
                sorter: (a, b) => {
                    const sum = (arr) => arr.reduce((acc, item) => acc + (item['logSize'] > 0 ? item['logSize'] : 0), 0);
                    return sum(a['replicas'] || []) - sum(b['replicas'] || []);
                },
                render: (y, record) => {
                    let totalLogSize = (record['replicas'] || []).reduce((acc, item) => acc + (item['logSize'] > 0 ? item['logSize'] : 0), 0);
                    if (totalLogSize < 0) {
                        return '不支持';
                    }
                    return renderSize(totalLogSize)
                }
        }, {
            title: 'Replicas',
            dataIndex: 'replicas',
            key: 'replicas',
            render: (replicas, record, index) => {
                return <Space>
                    {replicas.map(item => {
                        let logSize = ''
                        if (item['logSize'] > 0) {
                            logSize = renderSize(item['logSize']);
                        }
                        const host = item['host'] || '';
                        const port = item['port'] || '';
                        return <Tooltip key={'replicas-' + host}
                                        title={`${host}:${port} ${logSize}`}>
                            <Button size='small'>{item['id']}</Button>
                        </Tooltip>
                    })}
                </Space>;
            }
        }, {
            title: 'ISR',
            dataIndex: 'isr',
            key: 'isr',
            render: (isr, record, index) => {
                return <Space>
                    {isr.map(item => {
                        const host = item['host'] || '';
                        const port = item['port'] || '';
                        return <Tooltip key={'isr-' + host}
                                        title={`${host}:${port}`}>
                            <Button size='small'>{item['id']}</Button>
                        </Tooltip>
                    })}
                </Space>;
            }
        }, {
            title: 'Sync',
            dataIndex: 'partition',
            key: 'partition',
            render: (partition, record, index) => {
                return arrayEquals(record['replicas'], record['isr']).toString();
            }
        }];

        return (
            <div>
                <Table
                    rowKey='partition'
                    dataSource={this.state.items}
                    columns={columns}
                    position={'both'}
                    size={'middle'}
                    loading={this.state.loading}
                    pagination={{
                        showSizeChanger: true,
                        total: this.state.items.length,
                        showTotal: total => <FormattedMessage id="total-items" values={{total}}/>
                    }}
                />
            </div>
        );
    }
}

export default TopicPartition;
