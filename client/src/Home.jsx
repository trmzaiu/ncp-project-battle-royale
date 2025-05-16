import { useNavigate } from "react-router-dom";

const Home = () => {
  const navigate = useNavigate();

  return (
    <div className="flex items-center justify-center min-h-screen p-4">
      <button
        onClick={() => navigate("/auth")}
        className="bg-blue-500 text-white px-6 py-2 rounded hover:bg-blue-600"
      >
        Login
      </button>
    </div>
  );
};

export default Home;
