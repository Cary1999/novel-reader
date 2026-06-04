import { Navigate, useSearchParams } from "react-router-dom";

export function SearchPage() {
  const [searchParams] = useSearchParams();
  const target = searchParams.size ? `/?${searchParams}` : "/";
  return <Navigate to={target} replace />;
}
