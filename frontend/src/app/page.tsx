"use client"; // Mark this file as a Client Component

import React, { useEffect, useState } from "react";
import { useRouter } from "next/navigation"; // Use next/navigation instead of next/router
import { Component } from "../components/component"; // Messaging component
import { Login } from "../components/login"; // Import Login component

export default function Home() {
  const [currentUserName, setCurrentUserName] = useState<string | null>(null);
  const router = useRouter();

  useEffect(() => {
    // Check for a logged-in user (could be from localStorage, cookies, etc.)
    const storedUserName = localStorage.getItem("currentUserName");
    if (storedUserName) { 
      setCurrentUserName(storedUserName); // Set the user if found
    } else {
     // router.push("/");  //Redirect to login page if not found
    }
  }, [router]);

 /** if (!currentUserName) {
    // Show a loading state or just return null until the useEffect completes
    return null;
  }  **/ 

  return (
    <div className="grid grid-rows-[20px_1fr_20px] items-center justify-items-center min-h-screen p-8 pb-20 gap-16 sm:p-20 font-[family-name:var(--font-geist-sans)]">
      <main className="flex flex-col gap-8 row-start-2 items-center sm:items-start">
        {currentUserName ? (
          <Component currentUserName={currentUserName} /> // Pass the user name to the messaging component
        ) : (
          <Login /> // Render the login component if user is not logged in
        )}
      </main>
    </div>
  );
} 
