package example.net;

import java.util.ArrayList;
import java.util.List;

public record CsvRow(List<String> fields) {
    public CsvRow {
        fields = List.copyOf(fields);
    }

    public static CsvRow parse(String line) {
        List<String> fields = new ArrayList<>();
        StringBuilder current = new StringBuilder();
        boolean quoted = false;
        for (int index = 0; index < line.length(); index++) {
            char character = line.charAt(index);
            if (character == '"') {
                if (quoted && index + 1 < line.length() && line.charAt(index + 1) == '"') {
                    current.append('"');
                    index++;
                } else {
                    quoted = !quoted;
                }
            } else if (character == ',' && !quoted) {
                fields.add(current.toString());
                current.setLength(0);
            } else {
                current.append(character);
            }
        }
        if (quoted) throw new IllegalArgumentException("unterminated quoted field");
        fields.add(current.toString());
        return new CsvRow(fields);
    }
}
