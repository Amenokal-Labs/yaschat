import React, { useState, useEffect } from 'react';
import axios from 'axios';

const ChatComponent = () => {
    const [messages, setMessages] = useState([]);
    const [newMessage, setNewMessage] = useState('');
    const [from, setFrom] = useState('');
    const [to, setTo] = useState('');
    const [subject, setSubject] = useState('');

    useEffect(() => {
        if (from && to) {
            // Fetch messages when the component mounts and when 'from' or 'to' change
            fetchMessages();
        }
    }, [from, to]);

    const fetchMessages = async () => {
        try {
            const response = await axios.get(`http://localhost:8080/messages?from=${from}&to=${to}`);
            setMessages(response.data || []); // Set messages to an empty array if response.data is null
        } catch (error) {
            console.error('Error fetching messages:', error);
        }
    };

    const handleSendMessage = async () => {
        if (!newMessage.trim() || !from.trim() || !to.trim() || !subject.trim()) return;
    
        const messagePayload = {
            from: from,
            to: to,
            subject: subject,
            content: newMessage,
        };
    
        try {
            await axios.post('http://localhost:8080/message', messagePayload);
            setNewMessage('');
            fetchMessages();  // Refresh messages after sending a new one
        } catch (error) {
            console.error('Error sending message:', error);
        }
    };

    return (
        <div className="chat-container">
            <div className="input-area">
                <input 
                    type="text" 
                    value={from} 
                    onChange={(e) => setFrom(e.target.value)} 
                    placeholder="From" 
                />
                <input 
                    type="text" 
                    value={to} 
                    onChange={(e) => setTo(e.target.value)} 
                    placeholder="To" 
                />
                <input 
                    type="text" 
                    value={subject} 
                    onChange={(e) => setSubject(e.target.value)} 
                    placeholder="Subject" 
                />
                <input 
                    type="text" 
                    value={newMessage} 
                    onChange={(e) => setNewMessage(e.target.value)} 
                    placeholder="Type a message..." 
                />
                <button onClick={handleSendMessage}>Send</button>
            </div>
            <div className="messages">
                {messages && messages.length > 0 ? (
                    messages.map((message, index) => (
                        <div key={index} className="message">
                            <strong>From:</strong> {message.from} <br />
                            <strong>To:</strong> {message.to} <br />
                            <strong>Subject:</strong> {message.subject} <br />
                            <strong>Message:</strong> {message.content}
                        </div>
                    ))
                ) : (
                    <div>No messages available</div>
                )}
            </div>
        </div>
    );
};

export default ChatComponent;
