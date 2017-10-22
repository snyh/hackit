import React, { Component } from 'react';
import XTerm from "./react-xterm.js";

import { Table, Icon, Button, Message, Container, Grid, Header, Divider } from 'semantic-ui-react';
import { Link } from 'react-router-dom';

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
                        <Header>已连接到本地服务器 {this.localServer}</Header>
                        <ListConnection localServer={this.localServer}/>
                    </div>
                );
            case "offline":
                return (<div>无法与本地服务器连接，请使用客户端打开本页面</div>)
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

    render() {
        const rows = this.state.connections.map ( (v) =>  {
            return (
                <Table.Row key={v.UUID}>
                    <Table.Cell>
                        <Icon name='cloud' size='large' color={ v.Status === "running" ? 'green' : 'grey' } />
                        {v.Status}
                    </Table.Cell>
                    <Table.Cell>{v.UUID}</Table.Cell>
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
            </div>
        );
    }
}

class DetailView extends Component {
    render() {
        return (
            <Container>
                <Header>The terminal is operated by the hacker !</Header>
                <Divider/>
                <Grid divided>
                    <Grid.Row>
                        <Grid.Column height="600px">
                            <XTerm backend={this.props.backend} />
                        </Grid.Column>
                    </Grid.Row>
                    <Grid.Row>
                        <Grid.Column>
                            <Message>
                                <p><span>Hacker:</span> 123hhhhh chat message........</p>
                                <p><span>You:</span> hhhhh chat message........</p>
                            </Message>
                        </Grid.Column>
                    </Grid.Row>
                </Grid>
            </Container>
        );
    }
}

export default UserView;
