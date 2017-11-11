import React, { Component } from 'react';
import { Launcher } from 'react-chat-window';

class Chat extends Component {
    constructor(props) {
        super(props);
        this.state = {
            messageList: [],
            newMessagesCount: 0,
            isOpen: false,
        };


        if (props.ws) {
            console.log(props.ws)
            props.ws.onmessage = (e) => {
                const newMessagesCount = this.state.isOpen ? this.state.newMessagesCount : this.state.newMessagesCount + 1
                // receive from ws
                try {
                    const msg = JSON.parse(e.data)
                    this.setState({
                        newMessagesCount: newMessagesCount,
                        messageList: [...this.state.messageList, msg ]
                    })
                } catch (err) {
                    this.setState({
                        newMessagesCount: newMessagesCount,
                        messageList: [...this.state.messageList, {
                            author: "them",
                            type : "text",
                            data: {
                                text: `UNKNOWN MESSAGE ${e.data}`
                            }
                        }]
                    })
                }
            }
        }
    }

    _onMessageWasSent(message) {
        this.props.ws.send(JSON.stringify(message))
    }

    _handleClick() {
        this.setState({
            isOpen: !this.state.isOpen,
            newMessagesCount: 0
        })
    }

    render() {
        return (
            <div>
                <Launcher
                    agentProfile={{
                        teamName: 'react-live-chat',
                        imageUrl: 'https://a.slack-edge.com/66f9/img/avatars-teams/ava_0001-34.png'
                    }}
                    onMessageWasSent={this._onMessageWasSent.bind(this)}
                    messageList={this.state.messageList}
                    newMessagesCount={this.state.newMessagesCount}
                    handleClick={this._handleClick.bind(this)}
                    isOpen={this.state.isOpen}
                />
            </div>
        );
    }

}


export default Chat;
