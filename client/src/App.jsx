import { Routes, Route } from "react-router-dom";
import Home from "./Home";
import Game from "./pages/Game";
import Lobby from "./pages/Lobby";
import Auth from "./pages/Auth";
import PrivateRoute from "./routes/PrivateRoute";

function App() {
  return (
    <Routes>
      <Route path="/" element={<Home />} />
      <Route path="/auth" element={<Auth />} />
      <Route path="/lobby" element={<PrivateRoute><Lobby /></PrivateRoute>} />
      <Route path="/game" element={<PrivateRoute><Game /></PrivateRoute>} />
    </Routes>
  );
}

export default App;
