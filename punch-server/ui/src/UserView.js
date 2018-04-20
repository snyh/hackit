import React, { Component } from 'react';
import XTerm from "./react-xterm.js";

import { Table, Icon, Label, Button, Message, Container, Grid, Header, Divider } from 'semantic-ui-react';
import { Link } from 'react-router-dom';
import {CopyToClipboard} from 'react-copy-to-clipboard';

import { PleaseUseClient } from './Widget.js';

import Chat from './Chat.js';


class UserView extends Component {
    constructor(props) {
        super(props)

        const port = this.props.match.params.port;
        this.state = {
            localStatus: "unknown",
        }
        this.localServer = `127.0.0.1:${port}`
    }
    componentDidMount() {
        const tick = () => {
            fetch(`http://${this.localServer}/status`).then( (resp) => {
                resp.json().then( s => {
                    this.switchStatus(s)
                    this.timer = setTimeout(tick, 3000);
                })
            }).catch( () => {
                this.switchStatus("offline")
                this.timer = setTimeout(tick, 3000);
            })
        };
        tick();
    }
    componentWillUnmount() {
        clearTimeout(this.timer);
    }

    switchStatus(s) {
        if (["unknown","error", "online", "offline", "connected"].indexOf(s) === -1) {
            alert(`Invalid status. ${s}`)
        }
        this.setState({
            localStatus: s
        })
    }

    render() {
        switch (this.state.localStatus) {
            case "online":
                return (
                    <div>
                        <Header>已通过本机代理({this.localServer}) 连接到远程服务器({window.location.hostname}) </Header>
                        <ListConnection localServer={this.localServer}/>
                    </div>
                );
            case "offline":
                return <PleaseUseClient />;
            default:
                return (<div>loading {this.state.localStatus}</div>)
        }
    }
}

function compareHackItConn(a, b) {
    const ss = ["running", "readdy", "closed"]
    const cs = ss.indexOf(a.Status) - ss.indexOf(b.Status)
    if (cs !== 0) {
        return cs
    }
    return a.CreateAt < b.CreateAt
}

class ListConnection extends Component {
    state = {
        connections: [],
        msgType : "hidden"
    }
    componentDidMount() {
        const tick = () => {
            fetch(`http://${this.props.localServer}/listTTYs`).then( (resp) => {
                resp.json().then( s => {
                    s.sort( compareHackItConn)
                    this.setState({
                        connections: s ? s : []
                    })
                    this.timer = setTimeout(tick, 3000);
                })
            }).catch( () => {
                this.setState({
                        connections: []
                })
                this.timer = setTimeout(tick, 3000);
            })
        };
        tick();
    }
    componentWillUnmount() {
        clearTimeout(this.timer);
    }

    requestMagicKey = ()=>{
        const api = `http://${this.props.localServer}/requestTTY`
        fetch(api).then( (resp) => {
            resp.json().then( uuid => {
                this.setState({
                    msgType: "info",
                    msg: `a new magic key ${uuid} has been generated.`
                })
            })
        }).catch( (err) => {
            this.setState({
                msgType: "error",
                msg: `${err}`
            })
        })
    }

    changeDetail(uuid) {
        this.setState({
            detail: uuid,
        })
    }

    tryGetDetailWidget() {
        const uuid = this.state.detail
        if (uuid) {
            return <DetailView localServer={this.props.localServer} uuid={uuid} />
        }
        return null
    }

    render() {
        const rows = this.state.connections.map ( (v) =>  {
            const msg = `有个小问题需要你协助，请访问 ${window.location.origin}/connect/${v.UUID} 来远程操作我的系统吧。`;
//            const msg = `${window.location.origin}/connect/${v.UUID}`;
            return (
                <Table.Row key={v.UUID}>
                    <Table.Cell>
                        <Icon name='cloud' size='large' color={ v.Status !== "closed" ? 'green' : 'grey' } />
                        {v.Status}
                    </Table.Cell>
                    <Table.Cell>
                        {v.UUID}
                        <Label.Group size="mini">
                            <CopyToClipboard text={msg}
                                             onCopy={() => {this.setState({copied: true});alert("Magic Key 拷贝成功")}}>
                            <ActionButton action="copy" status={v.Status}
                                          content="Copy MagicKey" icon='copy' />
                            </CopyToClipboard>

                            <ActionButton action="see" status={v.Status}
                                          content="See" icon='camera retro' onClick={this.changeDetail.bind(this, v.UUID)}/>
                            <ActionButton action="delete" status={v.Status}
                                   content="delete" icon='trash outline'/>
                        </Label.Group>
                    </Table.Cell>
                    <Table.Cell>{(new Date(v.CreateAt)).toString()}</Table.Cell>
                </Table.Row>
            );
        })

        const msgType = {
            [this.state.msgType]: true,
        }

        return (
            <div>
                <Link to="/">Home</Link>

                <Table striped>
                    <Table.Header>
                        <Table.Row>
                            <Table.HeaderCell>状态</Table.HeaderCell>
                            <Table.HeaderCell>MagicKey</Table.HeaderCell>
                            <Table.HeaderCell>创建时间</Table.HeaderCell>
                        </Table.Row>
                    </Table.Header>
                    <Table.Body>
                        {rows}
                    </Table.Body>
                </Table>

                你还可以<Button onClick={this.requestMagicKey}>生成</Button>一个玩玩
                <Message {...msgType}>{this.state.msg}</Message>
                <Divider/>
                {this.tryGetDetailWidget()}
            </div>
        );
    }
}

class ActionButton extends Component {
    enabledFn = (action, status) => {
        const colors = {
            "copy": {
                "ready": true,
            },
            "see": {
                "running": true,
                "closed": true
            },
            "delete": {
                "ready": true,
                "running": true,
                "closed": true,
            }
        }
        if (colors[action] === undefined) {
            return false
        }
        return colors[action][status] === true
    }

    render() {
        const enabled  = this.enabledFn(this.props.action, this.props.status)
        if (enabled) {
            return <Label color="green" as='a' {...this.props} />
        } else {
            return <Label color="grey" {...this.props}/>
        }
    }
}

class DetailView extends Component {
    constructor(props) {
        super(props)
        this.state = {
            ttyWS: undefined,
            chatWS: undefined
        }
        const tty = new WebSocket(`ws://${props.localServer}/tty/${props.uuid}`)
        tty.onopen = this.setTTYWS.bind(this, tty)
        tty.onclose = this.setTTYWS.bind(this, undefined)

        const chat = new WebSocket(`ws://${props.localServer}/chat/${props.uuid}`)
        chat.onopen = this.setChatWS.bind(this, chat)
        chat.onclose = this.setChatWS.bind(this, undefined)

    }
    setTTYWS(b) {
        this.setState({
            ttyWS: b
        })
    }
    setChatWS(b) {
        this.setState({
            chatWS: b
        })
    }

    componentWillUnmount() {
        if (this.state.ttyWS) {
            this.state.ttyWS.close()
            this.setTTYWS(undefined)
        }
        if (this.state.chatWS) {
            this.state.chatWS.close()
            this.setChatWS(undefined)
        }
    }
    render() {
        const uuid = this.props.uuid;
        return (
            <Container>
            <Header>{uuid} 以下为远程实时操作视角，你只可以查看。或在直接切断连接。</Header>
            <Divider/>
            { this.state.ttyWS ? <XTerm backend={this.state.ttyWS} /> : <div>Connecting</div> }
            { this.state.chatWS ? <Chat ws={this.state.chatWS} /> : <div>Connecting</div> }
            </Container>
        );
    }
}

export default UserView;
