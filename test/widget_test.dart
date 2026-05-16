import 'package:flutter_test/flutter_test.dart';
import 'package:goalbooze/app.dart';

void main() {
  testWidgets('App should render home screen', (WidgetTester tester) async {
    await tester.pumpWidget(const GoalBoozeApp());
    expect(find.text('GoalBooze'), findsOneWidget);
  });
}
