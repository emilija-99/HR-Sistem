import { BrowserRouter, Routes, Route } from "react-router-dom";
import { routes } from "./routes";

export default function AppRouter() {
  return (
    <BrowserRouter>
      <Routes>
        {routes.map((r) => (
          <Route key={r.path} path={r.path} element={r.element} />
        ))}
      </Routes>
    </BrowserRouter>
  );
}
