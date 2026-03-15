import Button from "../ui/Button";
import { Card, CardContent, CardHeader } from "../ui/Card";
import EmptyState from "../ui/EmptyState";
import { ArrowUpRightIcon, BookmarkIcon, PlusIcon } from "../ui/Icons";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../ui/Table";

export default function BookmarksSection({ bookmarks, canManage = true, onAdd }) {
  return (
    <section id="bookmarks">
      <Card>
        <CardHeader
          action={
            <Button
              disabled={!canManage}
              leadingIcon={PlusIcon}
              onClick={onAdd}
              variant="secondary"
            >
              Add bookmark
            </Button>
          }
          description="Curated links for docs, dashboards, and tools that sit next to the discovered estate."
          title="Bookmarks"
        />
        <CardContent className="p-0">
          {bookmarks.length === 0 ? (
            <div className="px-5 py-5 sm:px-6">
              <EmptyState
                action={canManage ? onAdd : undefined}
                actionLabel="Add bookmark"
                body="Pin operator docs, third-party dashboards, or external tools used alongside your homelab."
                title="No bookmarks saved"
              />
            </div>
          ) : (
            <Table>
              <TableHead>
                <tr>
                  <TableHeader>Bookmark</TableHeader>
                  <TableHeader>Destination</TableHeader>
                  <TableHeader>Notes</TableHeader>
                  <TableHeader className="text-right">Open</TableHeader>
                </tr>
              </TableHead>
              <TableBody>
                {bookmarks.map((bookmark) => (
                  <TableRow key={bookmark.id}>
                    <TableCell className="min-w-[220px]">
                      <div className="flex items-center gap-3">
                        <span className="inline-flex h-9 w-9 items-center justify-center rounded-xl bg-slate-100 text-slate-600">
                          <BookmarkIcon className="h-4 w-4" />
                        </span>
                        <div>
                          <p className="font-medium text-slate-900">{bookmark.name}</p>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell className="max-w-[320px]">
                      <p className="truncate" title={bookmark.url}>
                        {bookmark.url}
                      </p>
                    </TableCell>
                    <TableCell className="max-w-[300px]">
                      <p className="truncate" title={bookmark.description || "No description"}>
                        {bookmark.description || "No description"}
                      </p>
                    </TableCell>
                    <TableCell className="text-right">
                      <Button
                        onClick={() =>
                          window.open(bookmark.url, "_blank", "noopener,noreferrer")
                        }
                        size="sm"
                        trailingIcon={ArrowUpRightIcon}
                        variant="ghost"
                      >
                        Open
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </section>
  );
}
