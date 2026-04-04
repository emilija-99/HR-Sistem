import Routes from "./routes";
import AuthProvider from "./providers/authProvider";
export default function App() {
  return (
    <AuthProvider>
      <Routes />;
    </AuthProvider>
  );
}
