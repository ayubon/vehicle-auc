import { useParams } from 'react-router-dom';

export default function AuctionDetailPage() {
  const { id } = useParams();
  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold">Auction #{id}</h1>
      <p className="text-muted-foreground">Auction details coming soon...</p>
    </div>
  );
}
