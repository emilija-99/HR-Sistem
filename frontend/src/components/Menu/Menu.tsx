import { Link } from "react-router-dom";

export default function Menu() {
  return (
    <nav style={{ padding: 20, borderBottom: "1px solid #ccc" }}>
      <Link to="/">Home </Link>
    </nav>
  );
}
