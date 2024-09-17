"use client";

import React, { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import Link from "next/link";

// Types for conversations and messages
type Message = {
  id: string;
  from: string;
  to: string;
  content: string;
  timestamp: string;
};

type Conversation = {
  conversation_id: string;
  participants: string[];
  last_message: Message;
};

// Main component function
export function Component({ currentUserName }: { currentUserName: string }) {
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [selectedConversation, setSelectedConversation] = useState<Conversation | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [newMessage, setNewMessage] = useState<string>("");

  useEffect(() => {
    if (currentUserName) {
      fetch(`http://localhost:8080/api/conversations?name=${currentUserName}`)
        .then((response) => response.json())
        .then((data) => setConversations(data))
        .catch((error) => console.error("Error fetching conversations:", error));
    }
  }, [currentUserName]);

  useEffect(() => {
    if (selectedConversation) {
      fetch(`http://localhost:8080/api/conversations/${selectedConversation.conversation_id}/messages`)
        .then((response) => response.json())
        .then((data) => setMessages(data))
        .catch((error) => console.error("Error fetching messages:", error));
    }
  }, [selectedConversation]);

  const getContactName = (participants: string[]): string|undefined => {
    return participants.find((name) => name !== currentUserName);
  };

  const createConversation = async (contactName: string) => {
    const participants = [currentUserName, contactName];
    const sortedParticipants = [...participants].sort();

    const newConversation = {
      participants: sortedParticipants,
    };

    try {
      const response = await fetch("http://localhost:8080/api/conversations", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(newConversation),
      });

      if (!response.ok) {
        const text = await response.text(); // Read the response as text
        throw new Error(`Failed to create conversation: ${text}`);
      }

      const newConv = await response.json();

      if (newConv && Array.isArray(conversations)) {
        setConversations((prevConversations) => [...(prevConversations || []), newConv]);
      }
    } catch (error) {
      console.error("Error creating conversation:", error);
    }
  };

  const handleNewMessageClick = () => {
    const contactName = prompt("Enter the name of the contact you want to message:");
    if (contactName) {
      createConversation(contactName);
    }
  };

  const handleSendMessage = async (e: React.FormEvent) => {
    e.preventDefault();
    if (newMessage.trim() === "" || !selectedConversation) return;

    const message = {
      from_name: currentUserName,
      to_name: getContactName(selectedConversation.participants),
      content: newMessage,
      timestamp: new Date().toISOString(),
    };

    try {
      const response = await fetch(`http://localhost:8080/api/conversations/${selectedConversation.conversation_id}/messages`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(message),
      });

      if (!response.ok) {
        throw new Error("Failed to send message");
      }

      setNewMessage("");
      setMessages((prevMessages) => [...prevMessages, message]);
    } catch (error) {
      console.error("Error sending message:", error);
    }
  };

  return (
    <div className="grid grid-cols-[300px_1fr] max-w-4xl w-full h-auto min-h-[500px] rounded-lg overflow-hidden border">
      <div className="bg-muted/20 p-3 border-r flex flex-col h-[80vh]"> {/* Smaller height */}
        <div className="flex items-center justify-between space-x-4">
          <div className="font-medium text-sm">Messenger</div>
          <Button variant="ghost" size="icon" className="rounded-full w-8 h-8" onClick={handleNewMessageClick}>
            <PenIcon className="h-4 w-4" />
            <span className="sr-only">New message</span>
          </Button>
        </div>
        <div className="py-4">
          <form>
            <Input placeholder="Search" className="h-8" />
          </form>
        </div>
        <div className="grid gap-2 flex-grow overflow-y-auto max-h-[300px]"> {/* Smaller conversation list height */}
          {conversations?.length > 0 ? (
            conversations.map((conversation) => (
              <Link
                key={conversation.conversation_id}
                href="#"
                className="flex items-center gap-4 p-2 rounded-lg hover:bg-muted/50 bg-muted"
                onClick={() => setSelectedConversation(conversation)}
              >
                <div className="grid gap-0.5">
                  <p className="text-sm font-medium leading-none">
                    {getContactName(conversation.participants)}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    {conversation.last_message.content} &middot; {new Date(conversation.last_message.timestamp).toLocaleTimeString()}
                  </p>
                </div>
              </Link>
            ))
          ) : (
            <p>No conversations found</p>
          )}
        </div>
      </div>
      <div className="flex flex-col h-[80vh]"> {/* Smaller messages section */}
        <div className="p-3 flex border-b items-center">
          {selectedConversation && (
            <div className="flex items-center gap-2">
              <div className="grid gap-0.5">
                <p className="text-sm font-medium leading-none">
                  {getContactName(selectedConversation.participants)}
                </p>
                <p className="text-xs text-muted-foreground">Active now</p>
              </div>
            </div>
          )}
        </div>
        <div className="grid gap-4 p-3 flex-grow overflow-y-auto max-h-[500px]">
          {messages
            .filter((message) => message.content.trim() !== "")
            .map((message) => (
              <div
                key={message.id}
                className={`flex w-max max-w-[65%] flex-col gap-1 px-3 py-2 text-sm rounded-lg shadow-sm ${
                  message.from === currentUserName
                    ? "ml-auto bg-blue-500 text-white" // Blue for your messages
                    : "bg-gray-300 text-black" // Grey for contact's messages
                }`}
                style={{ minHeight: "40px" }} // Ensure there's a small height for each message box
              >
                {message.content}
              </div>
            ))}
        </div>
        <div className="border-t">
          <form className="flex w-full items-center space-x-2 p-3" onSubmit={handleSendMessage}>
            <Input
              id="message"
              placeholder="Type your message..."
              className="flex-1"
              autoComplete="off"
              value={newMessage}
              onChange={(e) => setNewMessage(e.target.value)}
            />
            <Button type="submit" size="icon">
              <span className="sr-only">Send</span>
              <SendIcon className="h-4 w-4" />
            </Button>
          </form>
        </div>
      </div>
    </div>
  );
}

function PenIcon(props) {
  return (
    <svg
      {...props}
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M17 3a2.85 2.83 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z" />
    </svg>
  );
}

function SendIcon(props) {
  return (
    <svg
      {...props}
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="m22 2-7 20-4-9-9-4Z" />
      <path d="M22 2 11 13" />
    </svg>
  );
}
