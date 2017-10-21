import React, { Component } from 'react';
import './App.css';

import { Icon, Message, Confirm, Button, Loader, Container, Header, Divider } from 'semantic-ui-react';

import { Redirect, Route, Link, Switch } from 'react-router-dom';

import XTerm from './react-xterm.js';


const API_SERVER = "localhost:2207"

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
            history.replace("/");
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
        const backend = new WebSocket(`ws://${API_SERVER}/connect?uuid=${id}`)
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
                return (<div>链接已过期~~~请找小伙伴重新生成一份吧</div>)
            case "loading":
                return (<div>载入中~~~</div>)
            default:
                return (
                    <div>
                        Hello... try connecting to {id}
                        <XTerm backend={this.state.backend}>
                            <Loader content="连接中" />
                        </XTerm>
                    </div>
                )
        }
    }
}

class ListMagicLink extends Component {
    constructor(props) {
        super(props)
        this.state = {
            values : []
        }
    }

    componentDidMount() {
        const tick = () => {
            fetch(`http://${API_SERVER}/list`).then( (resp) => {
                resp.json().then( data => {
                    this.setState({values: data});
                    this.timer = setTimeout(tick, 3000);
                })
           })
        };
        tick();
    }

    componentWillUnmount() {
        clearTimeout(this.timer);
    }

    render() {
        const ids = this.state.values.map( id => {
            return <li key={id}><Link to={`/connect/${id}`}>{id}</Link></li>
        })
        return (
            <div>
                <Header> HackIt 管理后台 <Link to="/">Home</Link> </Header>
                <Divider/>
                <Header> 当前有 {ids.length} 个连接 </Header>
                <ul>{ids}</ul>
            </div>
        );
    }
}

class App extends Component {
    render() {
        return (
            <Container>
                <Switch>
                    <Route exact path="/" component={ListMagicLink}/>
                    <Route path="/connect/:id" component={MagicLinkWithEnsure}/>
                </Switch>
            </Container>
        );
    }
}

export default App;
