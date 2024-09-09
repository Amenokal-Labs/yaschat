import React, { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import Link from "next/link";
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar";

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
export function Component() {
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [selectedConversation, setSelectedConversation] = useState<Conversation | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [newMessage, setNewMessage] = useState<string>("");

  // Fetch conversations when the component mounts
  useEffect(() => {
    fetch("/api/conversations?user_id=currentUserId") // Replace 'currentUserId' with the actual user ID
      .then((response) => response.json())
      .then((data) => setConversations(data))
      .catch((error) => console.error("Error fetching conversations:", error));
  }, []);

  // Fetch messages when a conversation is selected
  useEffect(() => {
    if (selectedConversation) {
      fetch(`/api/conversations/${selectedConversation.conversation_id}/messages`)
        .then((response) => response.json())
        .then((data) => setMessages(data))
        .catch((error) => console.error("Error fetching messages:", error));
    }
  }, [selectedConversation]);

  // Handle sending a new message
  const handleSendMessage = async (e: React.FormEvent) => {
    e.preventDefault();
    if (newMessage.trim() === "" || !selectedConversation) return;

    const message = {
      from: "currentUserId", // Replace with the actual sender's ID
      to: selectedConversation.participants.find((id) => id !== "currentUserId"), // Find recipient
      content: newMessage,
      timestamp: new Date().toISOString(),
    };

    try {
      const response = await fetch(`/api/conversations/${selectedConversation.conversation_id}/messages`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(message),
      });

      if (!response.ok) {
        throw new Error("Failed to send message");
      }

      // Clear the input field and update messages state
      setNewMessage("");
      setMessages((prevMessages) => [...prevMessages, message]);
    } catch (error) {
      console.error("Error sending message:", error);
    }
  };

  return (
    <div className="grid grid-cols-[300px_1fr] max-w-4xl w-full rounded-lg overflow-hidden border">
      <div className="bg-muted/20 p-3 border-r">
        <div className="flex items-center justify-between space-x-4">
          <div className="font-medium text-sm">Messenger</div>
          <Button variant="ghost" size="icon" className="rounded-full w-8 h-8">
            <PenIcon className="h-4 w-4" />
            <span className="sr-only">New message</span>
          </Button>
        </div>
        <div className="py-4">
          <form>
            <Input placeholder="Search" className="h-8" />
          </form>
        </div>
        <div className="grid gap-2">
          {conversations.map((conversation) => (
            <Link
              key={conversation.conversation_id}
              href="#"
              className="flex items-center gap-4 p-2 rounded-lg hover:bg-muted/50 bg-muted"
              onClick={() => setSelectedConversation(conversation)}
            >
              <Avatar className="border w-10 h-10">
                <AvatarImage src="/placeholder-user.jpg" alt="Image" />
                <AvatarFallback>OM</AvatarFallback>
              </Avatar>
              <div className="grid gap-0.5">
                <p className="text-sm font-medium leading-none">
                  {conversation.participants.join(", ")}
                </p>
                <p className="text-xs text-muted-foreground">
                  {conversation.last_message.content} &middot; {new Date(conversation.last_message.timestamp).toLocaleTimeString()}
                </p>
              </div>
            </Link>
          ))}
        </div>
      </div>
      <div>
        <div className="p-3 flex border-b items-center">
          {selectedConversation && (
            <div className="flex items-center gap-2">
              <Avatar className="border w-10 h-10">
                <AvatarImage src="/placeholder-user.jpg" alt="Image" />
                <AvatarFallback>OM</AvatarFallback>
              </Avatar>
              <div className="grid gap-0.5">
                <p className="text-sm font-medium leading-none">
                  {selectedConversation.participants.join(", ")}
                </p>
                <p className="text-xs text-muted-foreground">Active now</p>
              </div>
            </div>
          )}
        </div>
        <div className="grid gap-4 p-3">
          {messages.map((message) => (
            <div key={message.id} className={`flex w-max max-w-[65%] flex-col gap-2 rounded-full px-4 py-2 text-sm ${message.from === "currentUserId" ? "ml-auto bg-primary text-primary-foreground" : "bg-muted"}`}>
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
  )
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
  )
}
