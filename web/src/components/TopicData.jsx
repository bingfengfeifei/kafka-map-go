import React, {Component} from 'react';
import {
    Row,
    Select,
    Form,
    Button,
    Typography,
    Tooltip,
    InputNumber,
    List,
    Space,
    Statistic, Col
} from "antd";
import request from "../common/request";
import qs from "qs";
import dayjs from "dayjs";

import {
    RightCircleTwoTone,
    DownCircleTwoTone
} from '@ant-design/icons';
import {Input} from "antd/lib/index";
import {FormattedMessage} from "react-intl";
import withRouter from "../hook/withRouter.jsx";
import {PageHeader} from "@ant-design/pro-components";

const {Text} = Typography;

class TopicData extends Component {

    form = React.createRef();

    state = {
        topic: undefined,
        clusterId: undefined,
        loading: false,
        items: [],
        topicInfo: undefined,
        offset: 0,
        partition: 0,
        count: 10,
        autoOffsetReset: 'newest'
    }

    componentDidMount() {
        let urlParams = new URLSearchParams(this.props.location.search);
        let clusterId = urlParams.get('clusterId');
        let topic = urlParams.get('topic');
        this.setState({
            clusterId: clusterId,
            topic: topic
        })
        this.loadTopicInfo(clusterId, topic);
    }

    loadTopicInfo = async (clusterId, topic) => {
        let response = await request.get(`/topics/${topic}?clusterId=${clusterId}`);
        let result = response && response.data ? response.data : { partitions: [] };
        this.setState({
            topicInfo: result
        }, this.handlePartitionChange);
    }

    handlePartitionChange = () => {
        const partitions = this.state.topicInfo && Array.isArray(this.state.topicInfo['partitions']) ? this.state.topicInfo['partitions'] : [];
        if (partitions.length === 0) {
            return;
        }
        const partitionIndex = this.state.partition < partitions.length ? this.state.partition : 0;
        if (partitionIndex !== this.state.partition) {
            this.setState({partition: partitionIndex});
        }
        const selectedPartition = partitions[partitionIndex];
        let endOffset = selectedPartition['endOffset'];
        let beginningOffset = selectedPartition['beginningOffset'];
        let offset = beginningOffset;
        if ('newest' === this.state.autoOffsetReset) {
            offset = endOffset - this.state.count;
            if (offset < beginningOffset) {
                offset = beginningOffset;
            }
        }
        this.setState({
            offset: offset
        })
        if (this.form.current) {
            this.form.current.setFieldsValue({'offset': offset});
        }
        this.handleReset();
    }

    handleReset = () => {
        this.setState({
            items: []
        })
    }

    pullMessage = async (queryParams) => {
        this.setState({
            loading: true
        })
        try {
            queryParams['clusterId'] = this.state.clusterId;
            let paramsStr = qs.stringify(queryParams);
            let response = await request.get(`/topics/${this.state.topic}/data?${paramsStr}`);
            let result = response && Array.isArray(response.data) ? response.data : [];
            // Auto-expand JSON by default
            result.forEach(item => {
                try {
                    let obj = JSON.parse(item['value']);
                    item['format'] = JSON.stringify(obj, null, 4);
                } catch (e) {
                    // Not valid JSON, keep as is
                }
            });
            this.setState({
                items: result
            })
        } finally {
            this.setState({
                loading: false
            })
        }

    }

    render() {

        const partitions = this.state.topicInfo && Array.isArray(this.state.topicInfo['partitions']) ?
            this.state.topicInfo['partitions'] : [];
        const partitionIndex = this.state.partition < partitions.length ? this.state.partition : 0;
        const selectedPartition = partitions.length > 0 ? partitions[partitionIndex] : null;

        return (
            <div>
                <div className='kd-page-header'>
                    <PageHeader
                        className="site-page-header"
                        onBack={() => {
                            this.props.navigate(-1);
                        }}
                        subTitle={<FormattedMessage id="consume-message"/>}
                        title={this.state.topic}
                    >
                        <Row>
                            <Space size='large'>
                                {
                                    this.state.topicInfo ?
                                        <>
                                            <Statistic title="Beginning Offset"
                                                       value={selectedPartition ? selectedPartition['beginningOffset'] : '-'}/>
                                            <Statistic title="End Offset"
                                                       value={selectedPartition ? selectedPartition['endOffset'] : '-'}/>
                                            <Statistic title="Size"
                                                       value={selectedPartition ? selectedPartition['endOffset'] - selectedPartition['beginningOffset'] : '-'}/>
                                        </>
                                        : undefined
                                }

                            </Space>
                        </Row>
                    </PageHeader>
                </div>

                <div className='kd-page-header' style={{padding: 20}}>
                    <Form ref={this.form} onFinish={this.pullMessage}
                          initialValues={{
                              count: this.state.count,
                              partition: this.state.partition,
                              autoOffsetReset: this.state.autoOffsetReset,
                          }}>
                        <Row gutter={24}>
                            <Col span={6}>
                                <Form.Item
                                    name={'partition'}
                                    label={'Partition'}
                                >
                                    <Select onChange={(value) => {
                                        this.setState({
                                            partition: value
                                        }, this.handlePartitionChange);
                                    }}>
                                        {
                                            partitions.map(item => {
                                                return <Select.Option key={'p' + item['partition']}
                                                                      value={item['partition']}>{item['partition']}</Select.Option>
                                            })
                                        }
                                    </Select>
                                </Form.Item>
                            </Col>

                            <Col span={6}>
                                <Form.Item
                                    name={'autoOffsetReset'}
                                    label={'Auto Offset Reset'}
                                >
                                    <Select onChange={(value) => {
                                        this.setState({
                                            autoOffsetReset: value
                                        }, this.handlePartitionChange);
                                    }}>
                                        <Select.Option value="earliest">
                                            <FormattedMessage id="earliest"/>
                                        </Select.Option>
                                        <Select.Option value="newest">
                                            <FormattedMessage id="newest"/>
                                        </Select.Option>
                                    </Select>
                                </Form.Item>
                            </Col>

                            <Col span={6}>
                                <Form.Item
                                    name={'offset'}
                                    label={'Offset'}
                                >
                                    {
                                        selectedPartition ?
                                            <InputNumber
                                                min={selectedPartition['beginningOffset']}
                                                max={selectedPartition['endOffset']}
                                                // defaultValue={this.state.topicInfo['partitions'][this.state.partition]['endOffset'] - this.state.count}
                                                value={this.state.offset}
                                                style={{width: '100%'}}
                                                onChange={(value) => {
                                                    this.setState({
                                                        offset: value
                                                    })
                                                }}
                                            />
                                            : undefined
                                    }

                                </Form.Item>
                            </Col>
                            <Col span={6}>
                                <Form.Item
                                    name={'count'}
                                    label={'Count'}
                                >
                                    <InputNumber min={1} style={{width: '100%'}}/>
                                </Form.Item>
                            </Col>

                            <Col span={6} key='keyFilter'>
                                <Form.Item
                                    name={'keyFilter'}
                                    label={'Key'}
                                >
                                    <Input allowClear placeholder="filter message key"/>
                                </Form.Item>

                            </Col>

                            <Col span={6} key='valueFilter'>
                                <Form.Item
                                    name={'valueFilter'}
                                    label={'Value'}
                                >
                                    <Input allowClear placeholder="filter message value"/>
                                </Form.Item>
                            </Col>

                            <Col span={6} key='jsonKey'>
                                <Form.Item
                                    name={'jsonKey'}
                                    label={'JSON Key'}
                                >
                                    <Input allowClear placeholder="JSON field name"/>
                                </Form.Item>

                            </Col>

                            <Col span={6} key='jsonValue'>
                                <Form.Item
                                    name={'jsonValue'}
                                    label={'JSON Value'}
                                >
                                    <Input allowClear placeholder="field value (fuzzy)"/>
                                </Form.Item>
                            </Col>
                            <Col span={12} style={{textAlign: 'right'}}>
                                <Space>
                                    <Button type="primary" htmlType="submit" loading={this.state.loading}>
                                        <FormattedMessage id="pull"/>
                                    </Button>

                                    <Button type="default" danger onClick={this.handleReset}>
                                        <FormattedMessage id="reset"/>
                                    </Button>
                                </Space>
                            </Col>
                        </Row>

                    </Form>
                </div>

                <div className='kd-content'>
                    <List
                        itemLayout="horizontal"
                        dataSource={this.state.items}
                        loading={this.state.loading}
                        pagination={{
                            showSizeChanger: true,
                            total: this.state.items.length,
                            showTotal: total => <FormattedMessage id="total-items" values={{total}}/>
                        }}
                        renderItem={(item, index) => {
                            const title = <>
                                <Space>
                                    <Text code>partition:</Text>
                                    <Text>{item['partition']}</Text>
                                    <Text code>key:</Text>
                                    <Text>{item['key']}</Text>
                                    <Text code>offset:</Text>
                                    <Text>{item['offset']}</Text>
                                    <Text code>timestamp:</Text>:
                                    <Tooltip
                                        title={dayjs(item['timestamp']).format("YYYY-MM-DD HH:mm:ss")}>
                                        <Text>{dayjs(item['timestamp']).fromNow()}</Text>
                                    </Tooltip>
                                </Space>
                            </>;

                            const description = <Row wrap={false}>
                                <Col flex="none">
                                    <div style={{padding: '0 5px'}}>
                                        {
                                            item['format'] ?
                                                <DownCircleTwoTone onClick={() => {
                                                    let items = this.state.items;
                                                    items[index]['format'] = undefined;
                                                    this.setState({
                                                        items: items
                                                    })
                                                }}/> :
                                                <RightCircleTwoTone onClick={() => {
                                                    let items = this.state.items;
                                                    try {
                                                        let obj = JSON.parse(items[index]['value']);
                                                        items[index]['format'] = JSON.stringify(obj, null, 4);
                                                        this.setState({
                                                            items: items
                                                        })
                                                    } catch (e) {

                                                    }
                                                }}/>
                                        }
                                    </div>
                                </Col>
                                <Col flex="auto">
                                    <pre>{item['format'] ? item['format'] : item['value']}</pre>
                                </Col>
                            </Row>;

                            return <List.Item>
                                <List.Item.Meta
                                    title={title}
                                    description={description}
                                />
                            </List.Item>;
                        }}
                    />
                </div>
            </div>
        );
    }

}

export default withRouter(TopicData);
