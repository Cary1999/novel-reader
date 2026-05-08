import { Search } from "lucide-react";
import { FormEvent, useState } from "react";
import type { Category } from "../api/types";

interface SearchFormProps {
  initialQuery?: string;
  initialCategory?: string;
  categories: Category[];
  onSubmit: (query: string, category: string) => void;
}

export function SearchForm({ initialQuery = "", initialCategory = "", categories, onSubmit }: SearchFormProps) {
  const [query, setQuery] = useState(initialQuery);
  const [category, setCategory] = useState(initialCategory);

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    onSubmit(query.trim(), category);
  }

  return (
    <form className="search-form" onSubmit={handleSubmit}>
      <label className="input-with-icon">
        <Search size={18} aria-hidden="true" />
        <span className="sr-only">搜索关键词</span>
        <input
          value={query}
          maxLength={80}
          onChange={(event) => setQuery(event.target.value)}
          placeholder="书名、作者或关键词"
        />
      </label>
      <label>
        <span className="sr-only">分类</span>
        <select value={category} onChange={(event) => setCategory(event.target.value)}>
          <option value="">全部分类</option>
          {categories.map((item) => (
            <option key={item.id} value={item.name}>{item.name}</option>
          ))}
        </select>
      </label>
      <button className="primary-button" type="submit">
        <Search size={17} aria-hidden="true" />
        搜索
      </button>
    </form>
  );
}
