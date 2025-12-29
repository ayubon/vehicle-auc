import { useParams } from 'react-router-dom';

export default function VehicleDetailPage() {
  const { id } = useParams();
  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold">Vehicle #{id}</h1>
      <p className="text-muted-foreground">Vehicle details coming soon...</p>
    </div>
  );
}
