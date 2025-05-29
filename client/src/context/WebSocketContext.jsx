// src/context/WebSocketContext.jsx
import React, { createContext, useContext, useEffect, useRef, useState } from "react";
import process from 'process'

const WebSocketContext = createContext(null);
const HOST = process.env.VITE_WS_URL;

export function WebSocketProvider({ children }) {
    const socketRef = useRef(null);
    const reconnectTimeout = useRef(null);
    const [isConnected, setIsConnected] = useState(false);
    const WS_URL = process.PROD
            ? "wss://royaka-2025.as.r.appspot.com/ws"
            : HOST || "ws://zang:8080/ws" || "ws://LAPTOPCUATUI:8080/ws" || "ws://192.168.1.4:8080/ws";


    // Store all onMessage callbacks to support multiple listeners
    const messageListeners = useRef(new Set());

    const connectWebSocket = React.useCallback(() => {
        socketRef.current = new WebSocket(WS_URL);

        socketRef.current.onopen = () => {
            console.log("[WS] Connected");
            setIsConnected(true);
        };

        socketRef.current.onmessage = (event) => {
            let message;
            try {
                message = JSON.parse(event.data);
            } catch {
                console.warn("[WS] Invalid JSON");
                return;
            }
            // Call all listeners with the message
            messageListeners.current.forEach((cb) => cb(message));
        };

        socketRef.current.onclose = () => {
            console.warn("[WS] Disconnected");
            setIsConnected(false);
            // Try reconnecting after 3s
            reconnectTimeout.current = setTimeout(() => {
                console.log("[WS] Reconnecting...");
                connectWebSocket();
            }, 3000);
        };

        socketRef.current.onerror = (err) => {
            console.error("[WS] Error:", err);
        };
    }, []);

    useEffect(() => {
        connectWebSocket();

        return () => {
            clearTimeout(reconnectTimeout.current);
            socketRef.current?.close();
        };
    }, [connectWebSocket]);

    // Function to send message if WS open
    const sendMessage = (msg) => {
        if (socketRef.current?.readyState === WebSocket.OPEN) {
            socketRef.current.send(JSON.stringify(msg));
        } else {
            console.warn("[WS] Not connected");
        }
    };

    // Function for components to subscribe to messages
    const subscribe = (callback) => {
        messageListeners.current.add(callback);
        // Return unsubscribe function
        return () => messageListeners.current.delete(callback);
    };

    const contextValue = React.useMemo(
        () => ({ sendMessage, subscribe, isConnected }),
        [isConnected]
    );

    return (
        <WebSocketContext.Provider value={contextValue}>
            {children}
        </WebSocketContext.Provider>
    );
}

// Hook for easier usage in components
export const useWebSocketContext = () => {
    return useContext(WebSocketContext);
};
