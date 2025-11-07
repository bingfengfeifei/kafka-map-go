import React, {Component} from 'react';
import {Button, Card, Form, Input, Typography} from "antd";
import './Login.css'
import request from "../common/request";
import {LockOutlined, UserOutlined} from '@ant-design/icons';

const {Title} = Typography;

class Login extends Component {

    state = {
        inLogin: false,
        height: window.innerHeight,
        width: window.innerWidth
    };

    componentDidMount() {
        window.addEventListener('resize', () => {
            this.setState({
                height: window.innerHeight,
                width: window.innerWidth
            })
        });
    }

    handleSubmit = async params => {
        this.setState({
            inLogin: true
        });

        try {
            let result = await request.post('/login', params);
            // 登录成功，保存token
            if (result.code === 200 && result.data && result.data.token) {
                localStorage.setItem('X-Auth-Token', result.data.token);
                // 跳转到首页
                window.location.href = "/"
            }
        } catch (error) {
            console.error('Login failed:', error);
            // 错误处理已在request.js中完成，这里不需要额外处理
        } finally {
            this.setState({
                inLogin: false
            });
        }
    };

    render() {
        return (
            <div className='login-bg'
                 style={{width: this.state.width, height: this.state.height, backgroundColor: '#F0F2F5'}}>
                <Card className='login-card' title={null}>
                    <div style={{textAlign: "center", margin: '15px auto 30px auto', color: '#1890ff'}}>
                        <Title level={1}>Kafka Map</Title>
                    </div>
                    <Form onFinish={this.handleSubmit} className="login-form">
                        <Form.Item name='username' rules={[{required: true, message: '请输入登录账号！'}]}>
                            <Input prefix={<UserOutlined/>} placeholder="登录账号"/>
                        </Form.Item>
                        <Form.Item name='password' rules={[{required: true, message: '请输入登录密码！'}]}>
                            <Input.Password prefix={<LockOutlined/>} placeholder="登录密码"/>
                        </Form.Item>
                        {/*<Form.Item name='remember' valuePropName='checked' initialValue={false}>*/}
                        {/*    <Checkbox>记住登录</Checkbox>*/}
                        {/*</Form.Item>*/}
                        <Form.Item>
                            <Button type="primary" htmlType="submit" className="login-form-button"
                                    loading={this.state.inLogin}>
                                登录
                            </Button>
                        </Form.Item>
                    </Form>
                </Card>
            </div>

        );
    }
}

export default Login;