import { Routes, Route } from "react-router-dom";
import Entry from "./pages/Entry";
import Game from "./pages/Game";
import Lobby from "./pages/Lobby";
import Auth from "./pages/Auth";
import PrivateRoute from "./routes/PrivateRoute";
import { WebSocketProvider } from "./context/WebSocketContext";

function App() {
    return (
        <WebSocketProvider>
            <Routes>
                <Route path="/" element={<Entry />} />
                <Route path="/auth" element={<Auth />} />
                <Route path="/lobby" element={<PrivateRoute><Lobby /></PrivateRoute>} />
                <Route path="/game" element={<PrivateRoute><Game /></PrivateRoute>} />
            </Routes>
        </WebSocketProvider>
    );
}

export default App;
