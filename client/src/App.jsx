import { Routes, Route } from "react-router-dom";
import Entry from "./pages/Entry";
import GameSimple from "./pages/GameSimple";
import GameEnhanced from "./pages/GameEnhanced";
import Lobby from "./pages/Lobby";
import Auth from "./pages/Auth";
import CardDesk from "./pages/CardDesk";
import PrivateRoute from "./routes/PrivateRoute";
import { WebSocketProvider } from "./context/WebSocketContext";

function App() {
    return (
        <WebSocketProvider>
            <Routes>
                <Route path="/" element={<Entry />} />
                <Route path="/auth" element={<Auth />} />
                <Route path="/lobby" element={<PrivateRoute><Lobby /></PrivateRoute>} />
                <Route path="/game-simple" element={<PrivateRoute><GameSimple /></PrivateRoute>} />
                <Route path="/game-enhanced" element={<PrivateRoute><GameEnhanced /></PrivateRoute>} />
                <Route path="/card-desk" element={<PrivateRoute><CardDesk /></PrivateRoute>} />
            </Routes>
        </WebSocketProvider>
    );
}

export default App;
