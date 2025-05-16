import { useEffect, useRef, useState } from "react";

export default function useWebSocket(onMessageHandler) {
    const [isConnected, setIsConnected] = useState(false);
    const socketRef = useRef(null);

    useEffect(() => {
        connect();

        return () => {
            if (socketRef.current && socketRef.current.readyState === WebSocket.OPEN) {
                socketRef.current.close();
            }
        };
    }, []);

    const connect = () => {
        const ws = new WebSocket("ws://localhost:8080/ws");

        ws.onopen = () => {
            console.log("✅ WebSocket connected");
            socketRef.current = ws;
            setIsConnected(true);
        };

        ws.onmessage = (event) => {
            const res = JSON.parse(event.data);
            onMessageHandler && onMessageHandler(res);
        };

        ws.onerror = (err) => {
            console.error("WebSocket error:", err);
        };

        ws.onclose = () => {
            console.warn("WebSocket closed. Reconnecting...");
            setIsConnected(false);
            setTimeout(connect, 3000);
        };
    };

    const sendMessage = (msg) => {
        if (socketRef.current && socketRef.current.readyState === WebSocket.OPEN) {
            socketRef.current.send(JSON.stringify(msg));
        } else {
            console.warn("WebSocket not connected yet.");
        }
    };

    return {
        sendMessage,
        isConnected,
    };
}
