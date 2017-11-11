import React, { Component } from 'react';
import './App.css';

import { Icon, Message, Confirm, Container, Header, Divider } from 'semantic-ui-react';

import { Route, Link, Switch } from 'react-router-dom';

import XTerm from './react-xterm.js';

import UserView from './UserView.js';

import { PleaseUseClient } from './Widget.js';

//const HTTP_SERVER = `${window.location.hostname}:2207`
const HTTP_SERVER = `${window.location.hostname}:8080`

class MagicLinkWithEnsure extends Component {
    state = {
        confirmed: false
    }
    handleConfirm = () => {
        this.setState({ confirmed: true })
    }
    handleCancel = () => {
        const history = this.props.history;

        if (history.length > 1) {
            history.goBack();
        } else {
            console.log(history);
            document.write("just close the window");
        }
    }
    render() {
        const id = this.props.match.params.id
        if (this.state.confirmed === true) {
            return <MagicLink magicKey={id}/>
        }
        const txt = (
            <Message>
                <Icon name="privacy" size="large" />
                这是个<b>一次性</b>连接。
                <p>
                    若你确定登录，其他人将无法再访问此连接。
                </p>
            </Message>
        );
        return (
            <Confirm
                header={`即将登录到 ${id}`}
                content={txt}
                open={true}
                onCancel={this.handleCancel}
                onConfirm={this.handleConfirm}
            />
        );
    }
}

class MagicLink extends Component {
    constructor(props) {
        super(props)
        this.state = {
            status: "loading",
        }

        const id = this.props.magicKey;
        const backend = new WebSocket(`ws://${HTTP_SERVER}/ws?uuid=${id}`)
        backend.onclose = this.handleError.bind(this)
        backend.onopen = this.handleOpen.bind(this, backend)
    }

    handleOpen(ws) {
        console.log("HHH>>>", ws)
        this.setState({
            status: "ok",
            backend: ws
        })
    }

    handleError() {
        this.setState({
            status: "error"
        })
    }

    render() {
        const id = this.props.magicKey
        switch(this.state.status) {
            case "error":
                return (<div>链接已过期~~~请找小伙伴重新生成一份吧 <Link to="/">Home</Link></div>)
            case "loading":
                return (<div>载入中~~~</div>)
            default:
                return (
                    <div>
                        Hello... try connecting to {id}  <Link to="/">Home</Link>
                        <XTerm backend={this.state.backend}></XTerm>
                    </div>
                )
        }
    }
}

class App extends Component {
    render() {
        return (
            <Container>
                <Switch>
                    <Route exact path="/" component={PleaseUseClient}/>
                    <Route path="/connect/:id" component={MagicLinkWithEnsure}/>
                    <Route path="/mysys/:port" component={UserView} />
                </Switch>
            </Container>
        );
    }
}

export default App;
