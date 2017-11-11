import React, { Component } from 'react';

import { Icon, Message, Confirm } from 'semantic-ui-react';

import XTerm from './react-xterm.js';

import Chat from './Chat.js';

import { Link } from 'react-router-dom';

import CONFIG from './config.js';

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
        const ttyWS = new WebSocket(`ws://${CONFIG.HTTP_SERVER}/connectTTY/${id}`)
        ttyWS.onclose = this.handleTTYError.bind(this)
        ttyWS.onopen = this.handleOpenTTY.bind(this, ttyWS)

        const chatWS = new WebSocket(`ws://${CONFIG.HTTP_SERVER}/connectChat/${id}`)
        chatWS.onclose = this.handleChatError.bind(this)
        chatWS.onopen = this.handleOpenChat.bind(this, chatWS)
    }

    handleOpenTTY(ws) {
        this.setState({
            status: "ok",
            ttyWS: ws
        })
    }
    handleOpenChat(ws) {
        this.setState({
            chatWS: ws
        })
    }

    handleChatError() {
        console.log("Chat Open ERROR")
    }
    handleTTYError() {
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
                        { this.state.ttyWS ? <XTerm backend={this.state.ttyWS} /> : <div>Connecting</div> }
                        { this.state.chatWS ? <Chat ws={this.state.chatWS} /> : <div>Connecting</div> }
                    </div>
                )
        }
    }
}


class MagicLinkDirectly extends Component {
    render() {
        const id = this.props.match.params.id
        return <MagicLink magicKey={id}/>
    }
}


export default MagicLinkDirectly
//export default MagicLinkWithEnsure
