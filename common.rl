%%{
machine common;

double_quote = '"';

escaped_unicode = '\\u' xdigit{4};

not_double_quote_or_escape = ([^"\\] - (0x00 .. 0x1f));
escaped_char = '\\' (["/\\bfnrt] | 'u' xdigit{4}) ;

json_space = [ \t\r\n];
json_true = 'true';
json_false = 'false';
json_null = 'null';

json_uint = [0] | [1-9][0-9]*;
json_int = '-'? json_uint;
json_number = json_int ('.'[0-9]+)? ([eE][+\-]?[0-9]+)?;

json_string = double_quote ( not_double_quote_or_escape | escaped_char )* double_quote;
json_bool = json_true | json_false;

}%%
